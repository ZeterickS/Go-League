package gametoimage

import (
	assethelper "discord-bot/internal/app/helper/assets"
	"discord-bot/types/match"
	"fmt"
	"image"
	"image/png"
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

		spellimage, _, err := image.Decode(spell)
		if err != nil {
			return nil, fmt.Errorf("failed to decode spell image: %w", err)
		}
		// Resize the spell image to 31x31
		resizedSpell := resize.Resize(32, 32, spellimage, resize.Lanczos3)
		spellFile, err := os.CreateTemp("", "resized_spell_*.png")
		if err != nil {
			return nil, fmt.Errorf("failed to create temp file for resized spell: %w", err)
		}

		err = png.Encode(spellFile, resizedSpell)
		if err != nil {
			return nil, fmt.Errorf("failed to encode resized spell to file: %w", err)
		}

		_, err = spellFile.Seek(0, 0)
		if err != nil {
			return nil, fmt.Errorf("failed to seek to beginning of spell file: %w", err)
		}

		err = builder.AddImage(spellFile, float64(i*32), 0, 32, 32)
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
		// Log the error and the first few bytes of the image file for debugging
		buf := make([]byte, 512)
		perk.Seek(0, 0) // Reset the reader to the beginning
		n, _ := perk.Read(buf)
		fmt.Printf("First %d bytes of the image: %x\n", n, buf[:n])
		fmt.Printf("Perk image format: %s\n", format)

		// Reset the reader to the beginning before adding the image
		perk.Seek(0, 0)

		perkimage, _, err := image.Decode(perk)
		if err != nil {
			return nil, fmt.Errorf("failed to decode perk image: %w", err)
		}
		// Resize the perk image to 30x30
		resizedPerk := resize.Resize(28, 28, perkimage, resize.Lanczos3)
		perkFile, err := os.CreateTemp("", "resized_perk_*.png")
		if err != nil {
			return nil, fmt.Errorf("failed to create temp file for resized perk: %w", err)
		}

		err = png.Encode(perkFile, resizedPerk)
		if err != nil {
			return nil, fmt.Errorf("failed to encode resized perk to file: %w", err)
		}

		_, err = perkFile.Seek(0, 0)
		if err != nil {
			return nil, fmt.Errorf("failed to seek to beginning of perk file: %w", err)
		}

		err = builder.AddImage(perkFile, float64((i*28)+2+i*2), 34, 28, 28)
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

func (ib *ImageBuilder) Save(outputPath string) error {
	return ib.Context.SavePNG(outputPath)
}
