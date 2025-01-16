package gametoimage

import (
	assethelper "discord-bot/internal/app/helper/assets"
	"discord-bot/types/match"
	"image"
	"os"

	"github.com/fogleman/gg"
	"github.com/nfnt/resize"
)

// ImageBuilder is the main structure for handling the template image and assets.
type ImageBuilder struct {
	Template image.Image
	Context  *gg.Context
}

// GameToImage generates an image based on the game data provided.
func GameToImage(participant match.Participant) (*os.File, error) {
	builder, err := NewImageBuilder()
	if err != nil {
		return nil, err
	}
	wd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	defaultImage, err := os.Open(wd + "/assets/15.1.1/template/template_empty.png")
	if err != nil {
		return nil, err
	}
	defer defaultImage.Close()

	// assemble from right to left

	itemImages, err := assethelper.GetItemFiles(participant.Items.ItemIDs)
	if err != nil {
		return nil, err
	}

	// Adding Items
	for i := 0; i < 6; i++ {
		if i < len(itemImages) {
			err = builder.AddImage(itemImages[i], float64(i*64+64), 0, 64, 64)
			if err != nil {
				return nil, err
			}
		} else {
			err = builder.AddImage(defaultImage, float64(i*64+64), 0, 64, 64)
			if err != nil {
				return nil, err
			}
		}
	}

	spellImages, err := assethelper.GetSpellFiles(participant.Spells.SpellIDs)
	if err != nil {
		return nil, err
	}

	// Adding Spells
	for i, spell := range spellImages {
		err = builder.AddImage(spell, float64(i*32), 0, 32, 32)
		if err != nil {
			return nil, err
		}
	}

	perkImages, err := assethelper.GetPerkFiles(participant.Perks)
	if err != nil {
		return nil, err
	}

	// Adding Perks
	for i, item := range perkImages {
		if i >= 2 {
			err = builder.AddImage(item, float64(i*32), 0, 32, 32)
			if err != nil {
				return nil, err
			}
		}
	}

	err = builder.Save("output.png")
	if err != nil {
		return nil, err
	}

	output, err := os.Open("output.png")
	if err != nil {
		return nil, err
	}

	return output, nil
}

func NewImageBuilder() (*ImageBuilder, error) {
	wd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	file, err := os.Open(wd + "/assets/15.1.1/template/template.png")
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

func (ib *ImageBuilder) AddImage(newfile *os.File, x, y, width, height float64) error {
	defer newfile.Close()

	asset, _, err := image.Decode(newfile)
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
