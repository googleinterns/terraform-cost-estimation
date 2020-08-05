package resources

//Resource is the interface of a general resource (ComputeInstance,...).
type Resouce interface {
	ExtractResource(jsonResourceInfo interface{})
	CompletePricingInfo()
	PrintPricingInfo()
}
