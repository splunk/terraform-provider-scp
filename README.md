# Terraform Provider for Splunk Cloud (Beta) 

This provider is currently in beta stage and only supports the index resource for Splunk Cloud deployments. 

## Requirements

-	[Terraform](https://www.terraform.io/downloads.html) >= 0.13.x
-	[Go](https://golang.org/doc/install) >= 1.18

## Building The Provider

1. Clone the repository
1. Create go ```src``` directory and setup ```$GOPATH ```
1. Enter the provider directory 
1. Compile the provider by running ```make build```

## Using the provider

- Install Terraform
- Build the binary by running ```make build```
- Initialize terraform by ```terraform init```
- To update run ```terraform plan``` first to check config diff
- Run ```terraform apply``` to apply configurations

## Contributions 

This provider is under development and is currently in beta. Currently, we are not accepting contributions, however, please use the Github Use issue tracker to submit bugs or request features! 

[Splunk Ideas](https://ideas.splunk.com/) is another place for your suggestions and [Splunk Answers](https://community.splunk.com/t5/Community/ct-p/en-us) for questions.




