package resources

type PricingInfo struct {
	ComponentUnit string //GiB, GB, MB
	HurlyUnitPrice int64 
	CurrencyType string //USD
	CurrancyUnit string  //"nano"
}

type CoreInfo struct {
	Type string
	Preemptible bool
	Number int
	Pricing PricingInfo
}

type MemoryInfo struct {
	Type string
	Preemptible bool
	Amount float64
    UnitType string
    Pricing PricingInfo
}

type ComputeInstance struct {
	Id string
	Name string
	MachineType string
	Regions []string

	Mem *MemoryInfo
	Cpu *CpuInfo
}

func (instance *ComputeInstace) GetResource() (jsonObject interface{}) {
}