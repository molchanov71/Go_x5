package _bonus

type Sale struct {
	Product          string
	Category         string
	Quantity         int
	UnitPriceKopecks int64
	Cancelled        bool
}

type Report struct {
	RevenueKopecks    int64
	UnitsSold         int
	RevenueByCategory map[string]int64
	TopProduct        string
	InvalidSales      int
}

func AnalyzeSales(sales []Sale) Report {
	var report Report
	report.RevenueByCategory = make(map[string]int64)
	quantities := make(map[string]int)
	topProduct := ""
	topQuantity := 0
	for _, sale := range sales {
		if sale.Cancelled {
			continue
		}
		if sale.Product == "" || sale.Category == "" || sale.Quantity <= 0 || sale.UnitPriceKopecks <= 0 {
			report.InvalidSales++
			continue
		}
		report.RevenueKopecks += sale.UnitPriceKopecks * int64(sale.Quantity)
		report.UnitsSold += sale.Quantity
		report.RevenueByCategory[sale.Category] += sale.UnitPriceKopecks * int64(sale.Quantity)
		quantities[sale.Product] += sale.Quantity
	}

	for product, quantity := range quantities {
		if quantity > topQuantity {
			topQuantity = quantity
			topProduct = product
		} else if quantity == topQuantity && topProduct > product {
			topProduct = product
		}
	}

	report.TopProduct = topProduct
	return report
}
