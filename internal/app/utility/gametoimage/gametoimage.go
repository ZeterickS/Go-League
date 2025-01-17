package gametoimage

import (
	assethelper "discord-bot/internal/app/helper/assets"
	"discord-bot/types/match"
	"fmt"
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

	defaultImagePath := wd + "/assets/15.1.1/template/template_empty.png"
	defaultImage, err := os.Open(defaultImagePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open default image: %w", err)
	}
	defer defaultImage.Close()

	// Log the default image path
	fmt.Printf("Default image path: %s\n", defaultImagePath)

	// assemble from right to left

	itemImages, err := assethelper.GetItemFiles(participant.Items.ItemIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to get item files: %w", err)
	}

	// Adding Items
	for i := 0; i < 6; i++ {
		if i < len(itemImages) {
			// Log the item image path
			fmt.Printf("Item image path: %s\n", itemImages[i].Name())

			err = builder.AddImage(itemImages[i], float64(i*64+64), 0, 64, 64)
			if err != nil {
				return nil, fmt.Errorf("failed to add item image: %w", err)
			}
		} else {
			err = builder.AddImage(defaultImage, float64(i*64+64), 0, 64, 64)
			if err != nil {
				return nil, fmt.Errorf("failed to add default image: %w", err)
			}
		}
	}

	spellImages, err := assethelper.GetSpellFiles(participant.Spells.SpellIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to get spell files: %w", err)
	}

	// Adding Spells
	for i, spell := range spellImages {
		// Log the spell image path
		fmt.Printf("Spell image path: %s\n", spell.Name())

		err = builder.AddImage(spell, float64(i*32), 0, 32, 32)
		if err != nil {
			return nil, fmt.Errorf("failed to add spell image: %w", err)
		}
	}

	perkImages, err := assethelper.GetPerkFiles(participant.Perks)
	if err != nil {
		return nil, fmt.Errorf("failed to get perk files: %w", err)
	}

	// Adding Perks
	for i, perk := range perkImages {
		// Log the perk image path
		fmt.Printf("Perk image path: %s\n", perk.Name())

		// Check the image format
		_, format, err := image.Decode(perk)
		if err != nil {
			return nil, fmt.Errorf("failed to decode perk image: %w", err)
		}
		fmt.Printf("Perk image format: %s\n", format)

		err = builder.AddImage(perk, float64(i*32), 32, 32, 32)
		if err != nil {
			return nil, fmt.Errorf("failed to add perk image: %w", err)
		}
	}

	err = builder.Save("output.png")
	if err != nil {
		return nil, fmt.Errorf("failed to save image: %w", err)
	}

	output, err := os.Open("output.png")
	if err != nil {
		return nil, fmt.Errorf("failed to open output image: %w", err)
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
