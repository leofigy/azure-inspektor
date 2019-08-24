package main


import(
	"fmt"
	"os"
	tf "github.com/hashicorp/terraform/terraform"
	ms "github.com/terraform-providers/terraform-provider-azurerm/azurerm"
)

func main(){
	provider := ms.Provider()
	resources := provider.Resources()

	fmt.Printf("Total resources\n", len(resources))

	// checking if we can get the schema
	if len(resources) < 1 {
		fmt.Println("No resources available")
		os.Exit(1)
	}

	resourceTypes := make([]string, len(resources))

	for i, r := range resources{
		resourceTypes[i] = r.Name
	}


	request := tf.ProviderSchemaRequest{
		ResourceTypes : resourceTypes,
	}

	schema, err := provider.GetSchema(&request)

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	for name, desc := range schema.ResourceTypes {
		fmt.Printf("%s:\n, %+v", name, desc)
	}
}
