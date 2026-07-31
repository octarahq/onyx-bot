package utils

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/disgoorg/disgo/discord"
	"github.com/disintegration/imaging"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

type CaptchaGoodAnswerIdx int

var questionsType = []string{"witchFruit", "number", "add"}

type CaptchaObject struct {
	Type string
	Name string
}

var possibleObjects = []CaptchaObject{
	CaptchaObject{
		Type: "fruits",
		Name: "apple",
	},
	CaptchaObject{
		Type: "fruits",
		Name: "banana",
	},
	CaptchaObject{
		Type: "fruits",
		Name: "strawberry",
	},
	CaptchaObject{
		Type: "fruits",
		Name: "citron",
	},
	CaptchaObject{
		Type: "vehicle",
		Name: "car",
	},
	CaptchaObject{
		Type: "vehicle",
		Name: "plane",
	},
	CaptchaObject{
		Type: "vehicle",
		Name: "bycle",
	},
	CaptchaObject{
		Type: "vehicle",
		Name: "boat",
	},
	CaptchaObject{
		Type: "pet",
		Name: "cat",
	},
	CaptchaObject{
		Type: "pet",
		Name: "duck",
	},
	CaptchaObject{
		Type: "pet",
		Name: "fish",
	},
}

func CaptchaBuildMessage(sessionId string, member discord.Member, guild discord.Guild) (discord.MessageCreate, CaptchaGoodAnswerIdx, error) {
	randomQuestionType := questionsType[rand.Intn(len(questionsType))]
	var question string
	var content []CaptchaObject
	var answers []string
	var goodIdx int
	randomQuestionType = "witchFruit"
	switch randomQuestionType {
	case "witchFruit":
		fruit := getRandomObjectFromCategory("fruits", 1)
		nbObstacles := 4 + rand.Intn(2)
		obstacles := getRandomObjectFromCategory("vehicle", nbObstacles)
		if len(fruit) == 0 || len(obstacles) == 0 {
			return discord.NewMessageCreateV2(), 0, fmt.Errorf("failed to build captcha: missing captcha objects")
		}
		goodIdx = rand.Intn(len(obstacles))
		for i := range len(obstacles) {
			var obj CaptchaObject
			if i == goodIdx {
				obj = fruit[0]
			} else {
				obj = obstacles[i]
			}
			content = append(content, obj)
			answers = append(answers, strconv.Itoa(i+1))
		}
		question = "Witch image is the **fruit** ?"
	}

	img := GenerateCaptchaImages(content, goodIdx)

	buf := new(bytes.Buffer)
	if err := png.Encode(buf, img); err != nil {
		return discord.NewMessageCreateV2(), 0, err
	}

	msg := discord.NewMessageCreateV2(
		discord.NewContainer(
			discord.NewSection(
				discord.NewTextDisplayf("## Welcome %s 👋", member.EffectiveName()),
				discord.NewTextDisplayf("Before you can access the **%s** server, we need to verify that you are not a robot.", guild.Name),
				discord.NewTextDisplayf("> You have **5 minutes** and **3 attempts** to solve the captcha.\n> \n> 💡 *If you fail or run out of time, you can click the \"Restart\" button. Once the captcha is solved, you will be granted access to the server.*"),
			).WithAccessory(discord.NewThumbnail(*guild.IconURL())),
			discord.NewTextDisplayf("**Question :** %s", question),
			discord.NewMediaGallery(
				discord.MediaGalleryItem{
					Media: discord.UnfurledMediaItem{
						URL: "attachment://captcha.png",
					},
				},
			),
		),
	).WithFiles(discord.NewFile("captcha.png", "The captcha image", buf))

	return msg, CaptchaGoodAnswerIdx(goodIdx), nil
}

func GenerateCaptchaImages(objects []CaptchaObject, goodIdx int) image.Image {
	width := 500
	height := 400
	img := imaging.New(width, height, color.NRGBA{43, 45, 49, 255})

	type objectPlacement struct {
		x       int
		y       int
		rotated *image.NRGBA
	}

	placements := make([]objectPlacement, len(objects))

	for i, o := range objects {
		objPath := fmt.Sprintf("data/images/security/captcha/%s.png", strings.ToLower(o.Name))
		objImg, err := imaging.Open(objPath)
		if err != nil {
			continue
		}

		objImg = imaging.Resize(objImg, 80, 0, imaging.Lanczos)

		angle := rand.Float64() * 360
		rotated := imaging.Rotate(objImg, angle, color.Transparent)
		placements[i] = objectPlacement{rotated: rotated}
	}

	goodRot := placements[goodIdx].rotated
	if goodRot != nil {
		maxX := width - goodRot.Bounds().Dx()
		maxY := height - goodRot.Bounds().Dy()
		if maxX <= 0 {
			maxX = 1
		}
		if maxY <= 0 {
			maxY = 1
		}
		placements[goodIdx].x = rand.Intn(maxX)
		placements[goodIdx].y = rand.Intn(maxY)
	}

	decoyIdx := (goodIdx + 1) % len(objects)

	for i := range objects {
		if i == goodIdx || placements[i].rotated == nil {
			continue
		}

		rotBounds := placements[i].rotated.Bounds()
		objW := rotBounds.Dx()
		objH := rotBounds.Dy()

		if i == decoyIdx && goodRot != nil {
			goodW := goodRot.Bounds().Dx()
			goodH := goodRot.Bounds().Dy()

			x := placements[goodIdx].x + goodW/2
			y := placements[goodIdx].y + goodH/2

			if x > width-objW {
				x = width - objW
			}
			if y > height-objH {
				y = height - objH
			}

			placements[i].x = x
			placements[i].y = y
		} else {
			maxX := width - objW
			maxY := height - objH
			if maxX <= 0 {
				maxX = 1
			}
			if maxY <= 0 {
				maxY = 1
			}
			placements[i].x = rand.Intn(maxX)
			placements[i].y = rand.Intn(maxY)
		}
	}

	for i, p := range placements {
		if p.rotated == nil {
			continue
		}

		img = imaging.Overlay(img, p.rotated, image.Pt(p.x, p.y), 1.0)

		indexText := fmt.Sprintf("%d", i+1)

		d := &font.Drawer{
			Dst:  img,
			Src:  image.NewUniform(color.RGBA{255, 255, 255, 255}),
			Face: basicfont.Face7x13,
			Dot: fixed.Point26_6{
				X: fixed.I(p.x + 5),
				Y: fixed.I(p.y + 15),
			},
		}
		d.DrawString(indexText)
	}

	drawNoise(img, 6, 150)

	return img
}

func drawNoise(img *image.NRGBA, numLines, numDots int) {
	bounds := img.Bounds()
	
	for i := 0; i < numDots; i++ {
		x, y := rand.Intn(bounds.Max.X), rand.Intn(bounds.Max.Y)
		col := color.RGBA{uint8(rand.Intn(256)), uint8(rand.Intn(256)), uint8(rand.Intn(256)), 255}
		img.Set(x, y, col)
		img.Set(x+1, y, col)
		img.Set(x, y+1, col)
		img.Set(x+1, y+1, col)
	}

	for i := 0; i < numLines; i++ {
		x0, y0 := rand.Intn(bounds.Max.X), rand.Intn(bounds.Max.Y)
		x1, y1 := rand.Intn(bounds.Max.X), rand.Intn(bounds.Max.Y)
		col := color.RGBA{uint8(rand.Intn(256)), uint8(rand.Intn(256)), uint8(rand.Intn(256)), 255}

		dx := x1 - x0
		if dx < 0 {
			dx = -dx
		}
		dy := y1 - y0
		if dy < 0 {
			dy = -dy
		}
		sx, sy := 1, 1
		if x0 >= x1 {
			sx = -1
		}
		if y0 >= y1 {
			sy = -1
		}
		err := dx - dy
		
		for {
			img.Set(x0, y0, col)
			img.Set(x0+1, y0, col)
			img.Set(x0, y0+1, col)
			
			if x0 == x1 && y0 == y1 {
				break
			}
			e2 := 2 * err
			if e2 > -dy {
				err -= dy
				x0 += sx
			}
			if e2 < dx {
				err += dx
				y0 += sy
			}
		}
	}
}

func getRandomObjectFromCategory(category string, times int) []CaptchaObject {
	var res []CaptchaObject

	filtered := make([]CaptchaObject, 0)
	for _, o := range possibleObjects {
		if o.Type == category {
			filtered = append(filtered, o)
		}
	}

	if len(filtered) == 0 {
		return res
	}

	for i := 0; i < times; i++ {
		res = append(res, filtered[rand.Intn(len(filtered))])
	}

	return res
}

func init() {
	rand.Seed(time.Now().UnixNano())
}
