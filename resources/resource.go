package resources

//Interface of a general resource (ComputeInstance,...)
type Resouce interface {
	ExtractResource(jsonObject interface{})
	CompletePricingInfo()
	PrintPricingInfo()
}

