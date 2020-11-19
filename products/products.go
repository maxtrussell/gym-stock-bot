package products

type Product struct {
	Name     string
	URL      string
	Brand    string
	Category string
	TestFile string
}

var Products = []Product{
	Product{
		Name:     "Rogue Olympic Plates",
		Brand:    "Rogue",
		Category: "multi",
		URL:      "https://www.roguefitness.com/rogue-olympic-plates",
	},
	Product{
		Name:     "Rogue Deep Dish Plates",
		Brand:    "Rogue",
		Category: "multi",
		URL:      "https://www.roguefitness.com/rogue-deep-dish-plates",
	},
	Product{
		Name:     "Rogue Steel Plates",
		Brand:    "Rogue",
		Category: "multi",
		URL:      "https://www.roguefitness.com/rogue-calibrated-lb-steel-plates",
	},
	Product{
		Name:     "Rogue HG Bumper Plates",
		Brand:    "Rogue",
		Category: "multi",
		URL:      "https://www.roguefitness.com/rogue-hg-2-0-bumper-plates",
	},
	Product{
		Name:     "Rogue Fleck Plates",
		Brand:    "Rogue",
		Category: "multi",
		URL:      "https://www.roguefitness.com/rogue-fleck-plates",
	},
	Product{
		Name:     "Rogue Echo Bumper Plate v2",
		Brand:    "Rogue",
		Category: "multi",
		URL:      "https://www.roguefitness.com/rogue-echo-bumper-plates-with-white-text",
	},
	Product{
		Name:     "Rogue Color Echo Bumper Plates",
		Brand:    "Rogue",
		Category: "multi",
		URL:      "https://www.roguefitness.com/rogue-color-echo-bumper-plate",
	},
	Product{
		Name:     "Rogue Mil Spec Bumper Plates",
		Brand:    "Rogue",
		Category: "multi",
		URL:      "https://www.roguefitness.com/rogue-us-mil-sprc-bumper-plates",
	},
	Product{
		Name:     "Rogue Mil Spec Echo Bumper Plates",
		Brand:    "Rogue",
		Category: "multi",
		URL:      "https://www.roguefitness.com/rogue-mil-echo-bumper-plates-black",
	},
	Product{
		Name:     "Rogue Ohio Power Bar Stainless Steel",
		Brand:    "Rogue",
		Category: "single",
		URL:      "https://www.roguefitness.com/rogue-45lb-ohio-power-bar-stainless",
	},
	Product{
		Name:     "Rep Fitness Iron Plates",
		Brand:    "RepFitness",
		Category: "rep",
		URL:      "https://www.repfitness.com/bars-plates/olympic-plates/iron-plates/rep-iron-plates",
	},
	Product{
		Name:     "Rep Fitness Black Bumper Plates",
		Brand:    "RepFitness",
		Category: "rep",
		URL:      "https://www.repfitness.com/bars-plates/olympic-plates/bumper-plates/rep-black-bumper-plates",
	},
	Product{
		Name:     "Rep Fitness Color Bumper Plates",
		Brand:    "RepFitness",
		Category: "rep",
		URL:      "https://www.repfitness.com/bars-plates/olympic-plates/rep-color-bumper-plates",
		TestFile: "test_pages/rep-color-bumper-plates.html",
	},
}

var TestProducts = []Product{
	Product{
		Name:     "Rep Fitness Color Bumper Plates",
		Brand:    "RepFitness",
		Category: "rep",
		URL:      "https://www.repfitness.com/bars-plates/olympic-plates/rep-color-bumper-plates",
		TestFile: "test_pages/rep-color-bumper-plates.html",
	},
	Product{
		Name:     "Rep Fitness Black Bumper Plates",
		Brand:    "RepFitness",
		Category: "rep",
		URL:      "https://www.repfitness.com/bars-plates/olympic-plates/bumper-plates/rep-black-bumper-plates",
		TestFile: "test_pages/rep-black-bumper-plates.html",
	},
	Product{
		Name:     "Rogue Color Echo Bumper Plates",
		Brand:    "Rogue",
		Category: "multi",
		URL:      "https://www.roguefitness.com/rogue-color-echo-bumper-plate",
		TestFile: "test_pages/rogue-color-echo-bumper-plate.html",
	},
	Product{
		Name:     "Rogue Ohio Power Bar Stainless Steel",
		Brand:    "Rogue",
		Category: "single",
		URL:      "https://www.roguefitness.com/rogue-45lb-ohio-power-bar-stainless",
		TestFile: "test_pages/rogue-45lb-ohio-power-bar-stainless.html",
	},
	Product{
		Name:     "Rogue Olympic Plates",
		Brand:    "Rogue",
		Category: "multi",
		URL:      "https://www.roguefitness.com/rogue-olympic-plates",
		TestFile: "test_pages/rogue-olympic-plates.html",
	},
	Product{
		Name:     "Rogue Fleck Plates",
		Brand:    "Rogue",
		Category: "multi",
		URL:      "https://www.roguefitness.com/rogue-fleck-plates",
		TestFile: "test_pages/rogue-fleck-plates.html",
	},
}
