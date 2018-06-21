package main

import (
	"github.com/satori/go.uuid"
	"code.cloudfoundry.org/lager"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/cloudfoundry-community/go-cfenv"
	"github.com/goji/httpauth"
	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/pivotal-cf/brokerapi"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
)

type KalturaInstances struct {
	Id          string `gorm:"primary_key;unique;not null"`
	PartnerId   int    `gorm:"not null"`
	AdminSecret string `gorm:"not null"`
}

type KalturaBroker struct {
	db *gorm.DB
}

func NewKalturaBroker() *KalturaBroker {
	db, err := gorm.Open("postgres", getDBParameters())
	if err != nil {
		panic(fmt.Sprintf("Failed to connect to the database: %v\n", err))
	}

	// Migrate the schema
	db.AutoMigrate(&KalturaInstances{})

	broker := &KalturaBroker{
		db: db,
	}
	return broker
}

func (b *KalturaBroker) Close() {
	if b.db != nil {
		db := b.db
		b.db = nil
		db.Close()
	}
}

func (b *KalturaBroker) Services(ctx context.Context) ([]brokerapi.Service, error) {
	log.Printf("Got a request to retrieve the catalog")

	return []brokerapi.Service{
		brokerapi.Service{
			ID:          uuid.NewV4(),
			Name:        "kaltura-vpaas",
			Description: "Use Kaltura to create Video Experiences and Workflows in your application",
			Bindable:    true,
			Metadata: &brokerapi.ServiceMetadata{
				DisplayName: "Video Platform as a Service - Kaltura",
				ImageUrl:    "https://vpaas.kaltura.com/images/VPaaS-logo-full.png",
				LongDescription: `
Kaltura VPaaS (Video Platform as a Service) allows you to build any video experience or workflow, and to integrate rich video experiences into existing applications, business workflows and environments.
Kaltura VPaaS eliminates all complexities involved in handling video at scale: ingestion, transcoding, metadata, playback, distribution, analytics, accessibility, monetization, security, search, interactivity and more.
Available as an open API, with a set of SDKs, developer tools and dozens of code recipes, we’re making the video experience creation process as easy as it gets.`,
				ProviderDisplayName: "Kaltura Inc.",
				DocumentationUrl:    "https://developer.kaltura.com",
				SupportUrl:          "https://forum.kaltura.org",
			},
			Plans: []brokerapi.ServicePlan{brokerapi.ServicePlan{
				ID:          uuid.NewV4(),
				Name:        "default",
				Description: "Pay As You Go with base REE package. For more details see: https://vpaas.kaltura.com/pricing",
			}},
		},
	}, nil
}

type ProvisionParameters struct {
	Name    string `json:"name"`
	Company string `json:"company"`
	Email   string `json:"email"`
}

type KalturaPartnerProvision struct {
	Id          int    `json:"id"`
	AdminSecret string `json:"adminSecret"`

	ObjectType   string `json:"objectType"`
	ErrorMessage string `json:"message"`
}

func (b *KalturaBroker) Provision(ctx context.Context, instanceID string, details brokerapi.ProvisionDetails, asyncAllowed bool) (brokerapi.ProvisionedServiceSpec, error) {
	log.Printf("Got a request to provision instanceId: %v\n", instanceID)
	retval := brokerapi.ProvisionedServiceSpec{
		DashboardURL: "https://kmc.kaltura.com/index.php/kmcng/login",
	}

	var params ProvisionParameters
	err := json.Unmarshal(details.RawParameters, &params)
	if err != nil {
		return retval, err
	}
	if params.Company == "" || params.Email == "" || params.Name == "" {
		return retval, errors.New("Missing parameters")
	}
	values := url.Values{}
	values.Add("partner[objectType]", "KalturaPartner")
	values.Add("partner[description]", "SAP Cloud Platform provisioned")
	values.Add("partner[name]", params.Company)
	values.Add("partner[adminName]", params.Name)
	values.Add("partner[adminEmail]", params.Email)
	values.Add("partner[referenceId]", instanceID)
	values.Add("format", "1")
	resp, err := http.PostForm("https://www.kaltura.com/api_v3/service/partner/action/register", values)
	if err != nil {
		return retval, err
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)

	log.Printf("Received a return value of %s\n", string(body))

	var kalturaResponse KalturaPartnerProvision
	err = json.Unmarshal(body, &kalturaResponse)
	if err != nil {
		return retval, err
	}
	if kalturaResponse.ObjectType == "KalturaAPIException" {
		return retval, errors.New(kalturaResponse.ErrorMessage)
	}
	log.Printf("Received a return value of %v\n", kalturaResponse)

	// Create
	b.db.Create(&KalturaInstances{
		Id:          instanceID,
		PartnerId:   kalturaResponse.Id,
		AdminSecret: kalturaResponse.AdminSecret,
	})

	return retval, nil
}

func (b *KalturaBroker) Deprovision(ctx context.Context, instanceID string, details brokerapi.DeprovisionDetails, asyncAllowed bool) (brokerapi.DeprovisionServiceSpec, error) {
	log.Printf("Got a request to deprovision a %v for instanceId: %v\n", details, instanceID)
	var record KalturaInstances
	b.db.First(&record, "Id = ?", instanceID)
	if record.Id == "" {
		return brokerapi.DeprovisionServiceSpec{}, errors.New("No such instance")
	}
	b.db.Delete(&record)
	return brokerapi.DeprovisionServiceSpec{}, nil
}

func (b *KalturaBroker) Bind(ctx context.Context, instanceID, bindingID string, details brokerapi.BindDetails) (brokerapi.Binding, error) {
	log.Printf("Got a request to bind bindingId %v for instanceId: %v\n", bindingID, instanceID)
	var record KalturaInstances
	b.db.First(&record, "Id = ?", instanceID)
	if record.Id == "" {
		return brokerapi.Binding{}, errors.New("No such instance")
	}
	return brokerapi.Binding{
		Credentials: map[string]interface{}{
			"adminSecret": record.AdminSecret,
			"partnerId":   record.PartnerId,
		},
	}, nil
}

func (b KalturaBroker) Unbind(ctx context.Context, instanceId, bindingID string, details brokerapi.UnbindDetails) error {
	log.Printf("Got a request to unbind bindingId %v for instanceId: %v\n", bindingID, instanceId)
	return nil
}

func (b *KalturaBroker) Update(ctx context.Context, instanceID string, details brokerapi.UpdateDetails, asyncAllowed bool) (brokerapi.UpdateServiceSpec, error) {
	return brokerapi.UpdateServiceSpec{}, nil
}

func (b *KalturaBroker) LastOperation(ctx context.Context, instanceID, operationData string) (brokerapi.LastOperation, error) {
	return brokerapi.LastOperation{}, nil
}

func getDBParameters() string {

	appEnv, err := cfenv.Current()
	if err != nil {
		panic(err)
	}
	pgServices, err := appEnv.Services.WithLabel("postgresql")
	if err != nil {
		panic(err)
	}
	if len(pgServices) != 1 {
		panic("Can't find the database")
	}
	return fmt.Sprintf("host=%s port=%s user=%s dbname=%s password=%s sslmode=disable",
		pgServices[0].Credentials["hostname"],
		pgServices[0].Credentials["port"],
		pgServices[0].Credentials["username"],
		pgServices[0].Credentials["dbname"],
		pgServices[0].Credentials["password"],
	)
}

func main() {

	router := mux.NewRouter().StrictSlash(true)
	brokerLogger := lager.NewLogger("broker")
	KalturaBroker := NewKalturaBroker()
	defer KalturaBroker.Close()

	brokerUser := os.Getenv("BROKER_USER")
	brokerPass := os.Getenv("BROKER_PASS")

	router.Use(httpauth.SimpleBasicAuth(brokerUser, brokerPass))
	brokerapi.AttachRoutes(router, KalturaBroker, brokerLogger)

	//add authentication for broker paths
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Fatal(http.ListenAndServe(":"+port, router))

}
