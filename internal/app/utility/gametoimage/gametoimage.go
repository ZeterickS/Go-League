package gametoimage

import (
	"image"
	"log"
	"os"

	"github.com/nfnt/resize"

	"github.com/fogleman/gg"
)

// ImageBuilder is the main structure for handling the template image and assets.
type ImageBuilder struct {
	Template image.Image
	Context  *gg.Context
}

// perks match.Perks, Items match.Items, SummonerSpell match.Spells
func gametoimage() {
	builder, err := NewImageBuilder()
	if err != nil {
		log.Fatalf("Error loading template: %v", err)
	}

	items := []string{
		"../../assets/15.1.1/items/1004.png",
		"../../assets/15.1.1/items/1006.png",
		"../../assets/15.1.1/items/1001.png",
		// Add more item paths as needed
	}

	defaultImage := "../../assets/template/template_empty.png"

	for i := 0; i < 6; i++ {
		item := defaultImage
		if i < len(items) && items[i] != "" {
			item = items[i]
		}
		err = builder.AddImage(item, float64(i*64+64), 0, 64, 64)
		if err != nil {
			log.Fatalf("Error adding image: %v", err)
		}
	}

	summonerSpells := []string{
		"../../assets/15.1.1/spells/SummonerFlash.png",
		"../../assets/15.1.1/spells/SummonerDot.png",
	}

	for i := 0; i < 2; i++ {
		spell := defaultImage
		if i < len(summonerSpells) && summonerSpells[i] != "" {
			spell = summonerSpells[i]
		}
		err = builder.AddImage(spell, float64(i*32), 0, 32, 32)
		if err != nil {
			log.Fatalf("Error adding summoner spell: %v", err)
		}
	}

	Perks := []string{
		"../../assets/15.1.1/spells/SummonerFlash.png",
		"../../assets/15.1.1/spells/SummonerDot.png",
	}

	for i := 0; i < 2; i++ {
		perk := defaultImage
		if i < len(Perks) && Perks[i] != "" {
			perk = Perks[i]
		}
		err = builder.AddImage(perk, float64(i*32), 32, 32, 32)
		if err != nil {
			log.Fatalf("Error adding summoner spell: %v", err)
		}
	}

	err = builder.Save("output.png")
	if err != nil {
		log.Fatalf("Error saving image: %v", err)
	}
}

func NewImageBuilder() (*ImageBuilder, error) {
	file, err := os.Open("../../assets/template/template.png")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	template, _, err := image.Decode(file)
	if err != nil {
		return nil, err
	}

	// Create a drawing context with the size of the template image.
	dc := gg.NewContextForImage(template)

	return &ImageBuilder{
		Template: template,
		Context:  dc,
	}, nil
}
func (ib *ImageBuilder) AddImage(assetPath string, x, y, width, height float64) error {
	file, err := os.Open(assetPath)
	if err != nil {
		return err
	}
	defer file.Close()

	asset, _, err := image.Decode(file)
	if err != nil {
		return err
	}
	// Resize the asset to the specified dimensions.
	resizedAsset := resize.Resize(uint(width), uint(height), asset, resize.Lanczos3)

	// Draw the resized asset onto the context.
	ib.Context.DrawImage(resizedAsset, int(x), int(y))
	return nil
}

func resizeImage(img image.Image, width, height int) image.Image {
	dc := gg.NewContext(width, height)
	dc.DrawImageAnchored(img, width/2, height/2, 0.5, 0.5)
	return dc.Image()
}

func (ib *ImageBuilder) Save(outputPath string) error {
	return ib.Context.SavePNG(outputPath)
}
