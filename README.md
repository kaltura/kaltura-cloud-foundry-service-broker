# Cloud Foundry Service Broker for Kaltura VPaaS

This is a [Cloud Foundry Service Broker](https://www.openservicebrokerapi.org/) for provisioning of Kaltura VPaaS accounts on Cloud Foundry based PaaS providers. CF PaaS based cloud vendors and cloud admins using Cloud Foundry can utilize this project to streamline accounts provisioning of Kaltura VPaaS accounts. 

This project is written in Go and is relies on the [Pivotal CF BrokerAPI](https://github.com/pivotal-cf/brokerapi) and [Go-CF Environment](https://github.com/cloudfoundry-community/go-cfenv) libraries.

This reference implementation uses postgress DB to store the accounts mapping. 

# How To

To run this project you will need to have access to a Cloud Foundry based PaaS. For example the [SAP Cloud Platform](https://cloudplatform.sap.com/index.html).  
You will also need to have [Go-lang](https://golang.org/) and [Go-dep](https://golang.github.io/dep/) installed, and ensure that `~/go/bin` is in your PATH.

* Run `go get github.com/kaltura/kaltura-cloud-foundry-service-broker`
* Run `cd ~/go/src/github.com/kaltura/kaltura-cloud-foundry-service-broker` (`~/go/` being the directory where go get is configured to download repos to)
* Edit `manifest.yml` and configure your desired username and password.
* Run `dep ensure`
* Create a postgress service for DB and name it postgress (e.g. `cf create-service postgresql v9.6-dev postgres`)
* Run `cf create-service-broker ...` to create a new service broker in your CF environment (for example: `cf create-service-broker kaltura-vpaas user pass https://kaltura-vpaas.cfapps.eu10.hana.ondemand.com --space-scoped`)
* Running `cf m` will now show a new `kaltura-vpaas` service available. Now you can use `cf create-service ...` to setup the service in your space, and then `cf bind-service ...` to bind the Kaltura VPaaS service to your CF app.

> Or use this code as basis reference example to implementing a new generic Kaltura service in your Cloud Foundry Cloud offering. (contact us at VPaaS@kaltura.com if you have any questions)

# How you can help (guidelines for contributors) 
Thank you for helping Kaltura grow! If you'd like to contribute please follow these steps:
* Use the repository issues tracker to report bugs or feature requests
* Read [Contributing Code to the Kaltura Platform](https://github.com/kaltura/platform-install-packages/blob/master/doc/Contributing-to-the-Kaltura-Platform.md)
* Sign the [Kaltura Contributor License Agreement](https://agentcontribs.kaltura.org/)

# Project TODO

* add sample app that uses the deployed service in an application (read credentials env vars)
* add input params validations (email, and such)
* document params for create
* document manifest.yml
* add better logging and better Error messages
* verify that when writing instances into postgres we do not override existing instance (shouldn't happen, but just in case)
* add OAUth login to KMC (Dashboard)
* add more info about metering and more plans (usage reporting, etc.) -- using SAP CP as reference implementation

# Where to get help
* Join the [Kaltura Community Forums](https://forum.kaltura.org/) to ask questions or start discussions
* Read the [Code of conduct](https://forum.kaltura.org/faq) and be patient and respectful

# Get in touch
You can learn more about Kaltura and start a free trial at: http://corp.kaltura.com    
Contact us via Twitter [@Kaltura](https://twitter.com/Kaltura) or email: community@kaltura.com  
We'd love to hear from you!

# License and Copyright Information
All code in this project is released under the [MIT license](https://github.com/kaltura/kaltura-cloud-foundry-service-broker/blob/master/LICENSE) unless a different license for a particular library is specified in the applicable library path.   

Copyright © Kaltura Inc and SAP SE. All rights reserved.   
Authors and contributors: [Lior Okman](https://github.com/liorokman/) and [Zohar Babin](https://github.com/zoharbabin). Also see [GitHub contributors list](https://github.com/kaltura/kaltura-cloud-foundry-service-broker/graphs/contributors).  

### Open Source Libraries
Review the [list of Open Source 3rd party libraries](open-source-libraries.md) used in this project.
