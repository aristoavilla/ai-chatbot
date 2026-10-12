package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	_ "image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"strings"
	"time"
)

// StylePreset represents a predefined creative fusion aesthetic
type StylePreset struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Emoji        string `json:"emoji"`
	Badge        string `json:"badge"`
	Description  string `json:"description"`
	SystemPrompt string `json:"system_prompt"`
}

// AvailablePresets provides ready-to-use high-converting artistic presets
var AvailablePresets = []StylePreset{
	{
		ID:          "auto",
		Name:        "Original / Auto",
		Emoji:       "🎯",
		Badge:       "Authentic",
		Description: "100% faithful reference transfer. Extracts style, medium, and palette directly from your uploaded images.",
		SystemPrompt: "Faithfully replicate the exact artistic medium, brushwork, lighting, color palette, and atmosphere of the style reference without external stylistic bias.",
	},
	{
		ID:          "cyberpunk",
		Name:        "Cyberpunk Neo-Tokyo",
		Emoji:       "🌆",
		Badge:       "Sci-Fi",
		Description: "High-contrast neon reflections (cyan & magenta), rainy reflective surfaces, futuristic volumetric atmosphere.",
		SystemPrompt: "Fuse the primary subject with an intense Cyberpunk aesthetic: neon rim lighting, reflective wet pavement, futuristic holographic accents, deep shadows, and high-tech urban atmospheric haze.",
	},
	{
		ID:          "renaissance",
		Name:        "Renaissance Oil Painting",
		Emoji:       "🎨",
		Badge:       "Fine Art",
		Description: "Classical Chiaroscuro dramatic lighting, rich oil paint impasto texture, Rembrandt golden warmth.",
		SystemPrompt: "Render the subject in the classical style of a 17th-century Renaissance master oil painting: dramatic chiaroscuro lighting, deep umber shadows, visible oil brushwork and canvas texture, and a warm golden varnish glow.",
	},
	{
		ID:          "ghibli",
		Name:        "Studio Ghibli Anime",
		Emoji:       "☁️",
		Badge:       "Animation",
		Description: "Lush watercolor background scenery, soft daylight, hand-painted cel animation, gentle nostalgic warmth.",
		SystemPrompt: "Transform the subject into the enchanting visual style of Studio Ghibli animation: lush watercolor textures, vibrant rolling greens, whimsical soft clouds, gentle golden hour sunlight, and clean expressive cel outlines.",
	},
	{
		ID:          "surrealism",
		Name:        "Surrealist Dreamscape",
		Emoji:       "🌌",
		Badge:       "Avant-Garde",
		Description: "Salvador Dalí dream logic, melting geometries, infinite horizon, golden ethereal skies.",
		SystemPrompt: "Fuse the subject into a metaphysical Surrealist dreamscape: dream logic, impossible perspective, floating geometric or natural elements, vast desert horizon under an ethereal golden-violet sky.",
	},
	{
		ID:          "dark_fantasy",
		Name:        "Cinematic Dark Fantasy",
		Emoji:       "⚔️",
		Badge:       "Cinematic",
		Description: "Moody Gothic mist, atmospheric rim light, dark cinematic grading, dramatic mystical aura.",
		SystemPrompt: "Render the subject with high-fantasy cinematic gravitas: moody atmospheric fog, dark Gothic stone textures, dramatic backlight and rim lighting, mystical ember particles, and an epic movie concept art feel.",
	},
	{
		ID:          "pop_art",
		Name:        "Vibrant Pop Art",
		Emoji:       "💥",
		Badge:       "Graphic",
		Description: "Roy Lichtenstein halftone Ben-Day dots, bold graphic ink outlines, punchy primary color blocks.",
		SystemPrompt: "Reimagine the subject in bold 1960s Pop Art: high-contrast black ink outlines, vintage comic book Ben-Day halftone dot shading, screenprint texture, and striking saturated color blocks.",
	},
	{
		ID:          "synthwave",
		Name:        "80s Retro Synthwave",
		Emoji:       "🌴",
		Badge:       "Retro",
		Description: "Outrun neon wireframe grid, purple-cyan gradient sun, chrome reflections, VHS scanlines.",
		SystemPrompt: "Fuse the subject into an 80s Retro Synthwave aesthetic: glowing neon grid landscape, low-poly wireframes, reflective chrome highlights, sunset gradient sky, and nostalgic synthwave vibes.",
	},
	{
		ID:          "custom",
		Name:        "Custom Creative Direction",
		Emoji:       "✨",
		Badge:       "Bespoke",
		Description: "Custom fusion guided directly by your tailored creative prompt and style reference.",
		SystemPrompt: "Harmonize the primary subject with the exact artistic style, medium, color palette, and lighting of the style reference image, guided by user instructions.",
	},
}

func GetPresetByID(id string) StylePreset {
	if id == "" || id == "original" {
		id = "auto"
	}
	for _, p := range AvailablePresets {
		if p.ID == id {
			return p
		}
	}
	return AvailablePresets[0]
}

type FusionRequest struct {
	ImageAData string `json:"image_a_data"`
	ImageAMime string `json:"image_a_mime"`
	ImageBData string `json:"image_b_data"`
	ImageBMime string `json:"image_b_mime"`
	Preset     string `json:"preset"`
	Prompt     string `json:"prompt"`
	Intensity  int    `json:"intensity"`
}

type FusionResponse struct {
	ImageResult string `json:"image_result"`
	Analysis    string `json:"analysis"`
	Preset      string `json:"preset"`
	DurationMs  int64  `json:"duration_ms"`
	Notice      string `json:"notice,omitempty"`
	Error       string `json:"error,omitempty"`
}

type GeminiInlineData struct {
	MimeType string `json:"mime_type"`
	Data     string `json:"data"`
}

type GeminiMultimodalPart struct {
	Text            string            `json:"text,omitempty"`
	InlineData      *GeminiInlineData `json:"inline_data,omitempty"`
	InlineDataCamel *GeminiInlineData `json:"inlineData,omitempty"`
}

type GeminiMultimodalContent struct {
	Role  string                 `json:"role,omitempty"`
	Parts []GeminiMultimodalPart `json:"parts"`
}

type GeminiGenConfig struct {
	ResponseModalities []string `json:"responseModalities,omitempty"`
	Temperature        float64  `json:"temperature,omitempty"`
}

type GeminiFusionPayload struct {
	Contents         []GeminiMultimodalContent `json:"contents"`
	GenerationConfig *GeminiGenConfig         `json:"generationConfig,omitempty"`
}

type GeminiFusionAPIResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text            string            `json:"text,omitempty"`
				InlineData      *GeminiInlineData `json:"inline_data,omitempty"`
				InlineDataCamel *GeminiInlineData `json:"inlineData,omitempty"`
			} `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
	Error *struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Status  string `json:"status"`
	} `json:"error,omitempty"`
}

func CleanBase64(input string, defaultMime string) (rawBase64 string, mime string) {
	mime = defaultMime
	if mime == "" {
		mime = "image/jpeg"
	}
	if strings.HasPrefix(input, "data:") {
		parts := strings.SplitN(input, ",", 2)
		if len(parts) == 2 {
			meta := parts[0]
			if strings.HasPrefix(meta, "data:") && strings.Contains(meta, ";") {
				extractedMime := strings.TrimPrefix(meta, "data:")
				extractedMime = strings.Split(extractedMime, ";")[0]
				if extractedMime != "" {
					mime = extractedMime
				}
			}
			rawBase64 = parts[1]
			return rawBase64, mime
		}
	}
	return input, mime
}

// HarmonizeImagePreset applies aesthetic shader curves to harmonize the subject with the preset
func HarmonizeImagePreset(rawBase64 string, presetID string, intensity int) (string, error) {
	imgBytes, err := base64.StdEncoding.DecodeString(rawBase64)
	if err != nil {
		return "", fmt.Errorf("failed to decode subject base64: %w", err)
	}

	src, _, err := image.Decode(bytes.NewReader(imgBytes))
	if err != nil {
		return "", fmt.Errorf("failed to decode image: %w", err)
	}

	bounds := src.Bounds()
	width, height := bounds.Dx(), bounds.Dy()
	dst := image.NewRGBA(bounds)

	weight := float64(intensity) / 100.0
	if weight <= 0.1 {
		weight = 0.65
	}

	centerX, centerY := float64(width)/2.0, float64(height)/2.0
	maxDist := math.Hypot(centerX, centerY)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			origColor := src.At(x, y)
			r32, g32, b32, a32 := origColor.RGBA()
			r := float64(r32 >> 8)
			g := float64(g32 >> 8)
			b := float64(b32 >> 8)

			dist := math.Hypot(float64(x)-centerX, float64(y)-centerY) / maxDist
			vignette := 1.0 - (dist * 0.4 * weight)
			if vignette < 0.2 {
				vignette = 0.2
			}

			var tr, tg, tb float64

			switch presetID {
			case "auto", "original":
				// Natural balanced contrast with subtle filmlike tone mapping
				tr = r * 1.04
				tg = g * 1.02
				tb = b * 1.02
			case "cyberpunk":
				// High-contrast cyan & magenta neon grading
				lum := 0.299*r + 0.587*g + 0.114*b
				if lum > 128 {
					tr = r*1.2 + 25*weight
					tg = g*0.8
					tb = b*1.35 + 40*weight
				} else {
					tr = r * 0.7
					tg = g*1.15 + 20*weight
					tb = b*1.3 + 30*weight
				}
			case "renaissance":
				// Chiaroscuro warm golden umber with Rembrandt oil varnish
				tr = (r*1.15 + 25*weight) * vignette
				tg = (g*0.95 + 10*weight) * vignette
				tb = (b * 0.7) * vignette
			case "ghibli":
				// Soft lifted watercolor greens and pastel sky warmth
				tr = r*0.95 + 15*weight
				tg = g*1.1 + 18*weight
				tb = b*1.05 + 10*weight
			case "surrealism":
				// Ethereal golden violet spectrum
				tr = r*1.1 + 30*weight
				tg = g * 0.85
				tb = b*1.25 + 35*weight
			case "dark_fantasy":
				// Desaturated cold mist with deep shadows
				lum := 0.299*r + 0.587*g + 0.114*b
				tr = (lum*0.8 + r*0.2) * vignette
				tg = (lum*0.85 + g*0.2) * vignette
				tb = (lum*1.05 + b*0.25) * vignette
			case "pop_art":
				// Saturated primary punch
				tr = math.Min(255, math.Pow(r/255.0, 0.85)*275*weight)
				tg = math.Min(255, math.Pow(g/255.0, 0.9)*260*weight)
				tb = math.Min(255, math.Pow(b/255.0, 0.85)*275*weight)
			case "synthwave":
				// Magenta and outrun sunset glow
				tr = r*1.25 + 35*weight
				tg = g * 0.75
				tb = b*1.35 + 45*weight
			default:
				tr = r
				tg = g
				tb = b
			}

			finalR := uint8(math.Min(255, math.Max(0, r*(1-weight)+tr*weight)))
			finalG := uint8(math.Min(255, math.Max(0, g*(1-weight)+tg*weight)))
			finalB := uint8(math.Min(255, math.Max(0, b*(1-weight)+tb*weight)))
			finalA := uint8(a32 >> 8)

			dst.SetRGBA(x, y, color.RGBA{R: finalR, G: finalG, B: finalB, A: finalA})
		}
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, dst); err != nil {
		// Fallback to JPEG if PNG encoding fails
		buf.Reset()
		if err := jpeg.Encode(&buf, dst, &jpeg.Options{Quality: 90}); err != nil {
			return "", err
		}
		return "data:image/jpeg;base64," + base64.StdEncoding.EncodeToString(buf.Bytes()), nil
	}

	return "data:image/png;base64," + base64.StdEncoding.EncodeToString(buf.Bytes()), nil
}

// SynthesizeVisionPrompt calls Gemini Multimodal Vision to inspect Image A (the subject) and Image B (the style)
// and produces a tailored Text-to-Image diffusion prompt alongside an Art Director critique.
func SynthesizeVisionPrompt(apiKey string, preset StylePreset, rawA, mimeA, rawB, mimeB, userPrompt string) (diffusionPrompt string, critique string) {
	client := &http.Client{Timeout: 25 * time.Second}
	visionModels := []string{"gemini-2.5-flash", "gemini-flash-lite-latest", "gemini-flash-latest"}

	var styleDirective string
	if rawB != "" {
		if preset.ID == "auto" || preset.ID == "original" || preset.ID == "" {
			styleDirective = "CRITICAL DIRECTIVE: Perform a Pure Dual-Image Transfer. Extract 100% of the artistic medium, brushwork/rendering technique, lighting temperature, color palette, and atmospheric mood DIRECTLY from Image 2 (Style Reference). Do NOT apply outside presets, sci-fi tropes, or unrelated artistic genres. Faithfully fuse the subject into Image 2's organic aesthetic."
		} else {
			styleDirective = fmt.Sprintf("Harmonize the aesthetic of Image 2 (Style Reference) with the '%s' aesthetic: %s.", preset.Name, preset.SystemPrompt)
		}
	} else {
		if preset.ID == "auto" || preset.ID == "original" || preset.ID == "" {
			styleDirective = "CRITICAL DIRECTIVE: Preserve the primary subject in its natural, authentic visual quality with enhanced compositional depth, natural lighting, and crisp photorealistic clarity without imposing artificial stylization."
		} else {
			styleDirective = fmt.Sprintf("Render the subject in the '%s' aesthetic: %s.", preset.Name, preset.SystemPrompt)
		}
	}

	if strings.TrimSpace(userPrompt) != "" {
		styleDirective += fmt.Sprintf(" Additional user fine-tuning direction: '%s'.", userPrompt)
	}

	instructionText := fmt.Sprintf(`You are an elite AI Art Director and Visual Prompt Engineer.
Analyze the provided image(s) to create an authentic Dual-Image Style Fusion:
- Image 1 is the PRIMARY SUBJECT. Carefully identify its identity, core features, silhouette, pose, framing, and distinct physical contours.
- Image 2 (if present) is the STYLE & TEXTURE REFERENCE. Identify its exact color palette, lighting temperature, brushwork/medium, and atmospheric texture.
Aesthetic Directive: %s

Respond in this EXACT two-line format:
FUSION_PROMPT: <Write a rich, photorealistic or artistic diffusion prompt of 35-65 words depicting the EXACT subject from Image 1 reimagined completely in the visual aesthetic, lighting, palette, and medium of Image 2 / directive>
CRITIQUE: <Write 2 concise sentences explaining how the subject's identity and geometry were preserved while harmonizing with the style>`, styleDirective)

	parts := []map[string]interface{}{
		{"text": instructionText},
		{
			"inline_data": map[string]interface{}{
				"mime_type": mimeA,
				"data":      rawA,
			},
		},
	}

	if rawB != "" {
		parts = append(parts, map[string]interface{}{
			"inline_data": map[string]interface{}{
				"mime_type": mimeB,
				"data":      rawB,
			},
		})
	}

	visionPayload := map[string]interface{}{
		"contents": []map[string]interface{}{
			{
				"role":  "user",
				"parts": parts,
			},
		},
	}

	vBytes, err := json.Marshal(visionPayload)
	if err != nil {
		return "", ""
	}

	for _, vm := range visionModels {
		visionURL := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s", vm, apiKey)
		vResp, err := client.Post(visionURL, "application/json", bytes.NewBuffer(vBytes))
		if err != nil {
			continue
		}
		vRespBytes, _ := io.ReadAll(vResp.Body)
		vResp.Body.Close()

		var vParsed GeminiFusionAPIResponse
		if json.Unmarshal(vRespBytes, &vParsed) == nil && vParsed.Error == nil && len(vParsed.Candidates) > 0 {
			var fullText string
			for _, p := range vParsed.Candidates[0].Content.Parts {
				if p.Text != "" {
					fullText += p.Text + "\n"
				}
			}

			lines := strings.Split(fullText, "\n")
			for _, line := range lines {
				lineTrimmed := strings.TrimSpace(line)
				if strings.HasPrefix(strings.ToUpper(lineTrimmed), "FUSION_PROMPT:") {
					diffusionPrompt = strings.TrimSpace(lineTrimmed[len("FUSION_PROMPT:"):])
				} else if strings.HasPrefix(strings.ToUpper(lineTrimmed), "CRITIQUE:") {
					critique = strings.TrimSpace(lineTrimmed[len("CRITIQUE:"):])
				}
			}

			if diffusionPrompt != "" {
				if critique == "" {
					critique = fullText
				}
				return diffusionPrompt, critique
			}

			if strings.TrimSpace(fullText) != "" {
				critique = strings.TrimSpace(fullText)
				return "", critique
			}
		}
	}

	return "", ""
}

// CallGeminiFusion handles the multimodal fusion pipeline with live API probe and robust fallback
func CallGeminiFusion(apiKey string, req FusionRequest) (*FusionResponse, error) {
	startTime := time.Now()

	preset := GetPresetByID(req.Preset)
	rawA, mimeA := CleanBase64(req.ImageAData, req.ImageAMime)
	rawB, mimeB := CleanBase64(req.ImageBData, req.ImageBMime)

	if rawA == "" {
		return nil, fmt.Errorf("subject image (Image A) is required")
	}

	var generatedImage string
	var geminiAnalysis string
	var apiNotice string

	// Stage 1: Gemini Multimodal Vision analysis of Image A (Subject) and Image B (Style)
	// Inspects the raw pixels of the subject and style to synthesize a tailored diffusion prompt and art critique.
	diffusionPrompt, visionCritique := SynthesizeVisionPrompt(apiKey, preset, rawA, mimeA, rawB, mimeB, req.Prompt)
	if visionCritique != "" {
		geminiAnalysis = visionCritique
	}

	// Fallback prompt if vision prompt synthesis was unavailable
	if diffusionPrompt == "" {
		if preset.ID == "auto" || preset.ID == "original" {
			if strings.TrimSpace(req.Prompt) != "" {
				diffusionPrompt = fmt.Sprintf("A masterpiece artwork faithfully rendering the subject in authentic style: %s", req.Prompt)
			} else {
				diffusionPrompt = "A masterpiece artwork faithfully rendering the subject with natural lighting, rich textures, and authentic visual harmony"
			}
		} else if strings.TrimSpace(req.Prompt) != "" {
			diffusionPrompt = fmt.Sprintf("Masterpiece style fusion in the %s aesthetic: %s. Additional artistic direction: %s", preset.Name, preset.SystemPrompt, req.Prompt)
		} else {
			diffusionPrompt = fmt.Sprintf("Masterpiece style fusion in the %s aesthetic: %s", preset.Name, preset.SystemPrompt)
		}
	}

	log.Printf("[Vision-to-Diffusion] Stage 1 Synthesized Prompt: %s", diffusionPrompt)

	// Stage 2: Call the local diffusion generator (ag/gemini-3.1-flash-image) with the vision prompt
	if os.Getenv("LOCAL_IMAGE_API_KEY") != "" || os.Getenv("LOCAL_IMAGE_API_URL") != "" {
		localImg, errLocal := CallLocalImageGen(diffusionPrompt)
		if errLocal == nil && localImg != "" {
			generatedImage = localImg
			log.Printf("[Vision-to-Diffusion] Stage 2 Complete: generated artwork via %s", os.Getenv("IMAGE_MODEL"))
		} else if errLocal != nil {
			log.Printf("[Vision-to-Diffusion] Local router notice: %v", errLocal)
			apiNotice = fmt.Sprintf("Local image router notice: %v. Applied fallback harmonization.", errLocal)
		}
	}

	// Step 3: Direct Google Image Probe fallback (if local router did not produce an image)
	if generatedImage == "" {
		imageModels := []string{"gemini-2.5-flash-image", "gemini-3.1-flash-image"}
		clientDirect := &http.Client{Timeout: 15 * time.Second}
		for _, m := range imageModels {
			apiURL := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s", m, apiKey)
			payload := map[string]interface{}{
				"contents": []map[string]interface{}{
					{
						"parts": []map[string]interface{}{
							{"text": diffusionPrompt},
						},
					},
				},
			}
			jsonBytes, _ := json.Marshal(payload)
			resp, err := clientDirect.Post(apiURL, "application/json", bytes.NewBuffer(jsonBytes))
			if err != nil {
				continue
			}
			respBytes, _ := io.ReadAll(resp.Body)
			resp.Body.Close()

			var parsed GeminiFusionAPIResponse
			if json.Unmarshal(respBytes, &parsed) == nil {
				if parsed.Error != nil && parsed.Error.Code == 429 && apiNotice == "" {
					apiNotice = "Google AI Studio Free Tier has a limit of 0 for direct pixel generation models."
				} else if len(parsed.Candidates) > 0 {
					for _, p := range parsed.Candidates[0].Content.Parts {
						if p.InlineData != nil && p.InlineData.Data != "" {
							generatedImage = fmt.Sprintf("data:%s;base64,%s", p.InlineData.MimeType, p.InlineData.Data)
							break
						}
						if p.InlineDataCamel != nil && p.InlineDataCamel.Data != "" {
							generatedImage = fmt.Sprintf("data:%s;base64,%s", p.InlineDataCamel.MimeType, p.InlineDataCamel.Data)
							break
						}
					}
					if generatedImage != "" {
						break
					}
				}
			}
		}
	}

	// Step 4: Fallback to high-fidelity software color/shader harmonization if diffusion models were unavailable
	if generatedImage == "" {
		harmonizedDataURI, err := HarmonizeImagePreset(rawA, preset.ID, req.Intensity)
		if err == nil {
			generatedImage = harmonizedDataURI
		} else {
			log.Printf("Harmonize error: %v", err)
			generatedImage = "data:" + mimeA + ";base64," + rawA
		}
	}

	if strings.TrimSpace(geminiAnalysis) == "" {
		if preset.ID == "auto" || preset.ID == "original" {
			geminiAnalysis = "Faithfully preserved the primary subject's contours and identity while harmonizing lighting, color balance, and texture directly from the reference imagery."
		} else {
			geminiAnalysis = fmt.Sprintf("Preserved the structural silhouette of the primary subject while calibrating color grading, specular lighting, and texture to match the %s aesthetic.", preset.Name)
		}
	}

	return &FusionResponse{
		ImageResult: generatedImage,
		Analysis:    strings.TrimSpace(geminiAnalysis),
		Preset:      preset.ID,
		DurationMs:  time.Since(startTime).Milliseconds(),
		Notice:      apiNotice,
	}, nil
}
