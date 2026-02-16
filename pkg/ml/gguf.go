package ml

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// GGUF format constants.
const (
	ggufMagic     = 0x46554747 // "GGUF" little-endian
	ggufVersion   = 3
	ggufAlignment = 32
)

// GGUF metadata value types.
const (
	ggufTypeUint32  = 4
	ggufTypeFloat32 = 6
	ggufTypeString  = 8
)

// GGML tensor data types.
const (
	ggmlTypeF32  = 0
	ggmlTypeF16  = 1
	ggmlTypeBF16 = 30
)

// ggufMetadata is a key-value pair in the GGUF header.
type ggufMetadata struct {
	key       string
	valueType uint32
	value     interface{} // string, uint32, or float32
}

// ggufTensor describes a tensor in the GGUF file.
type ggufTensor struct {
	name  string
	dims  []uint64
	dtype uint32 // ggmlType*
	data  []byte
}

// gemma3ModuleMap maps HuggingFace module names to GGUF tensor names.
var gemma3ModuleMap = map[string]string{
	"self_attn.q_proj": "attn_q",
	"self_attn.k_proj": "attn_k",
	"self_attn.v_proj": "attn_v",
	"self_attn.o_proj": "attn_output",
	"mlp.gate_proj":    "ffn_gate",
	"mlp.up_proj":      "ffn_up",
	"mlp.down_proj":    "ffn_down",
}

var mlxLoraKeyRe = regexp.MustCompile(`^model\.layers\.(\d+)\.(.*?)\.(lora_[ab])$`)

// MLXTensorToGGUF converts an MLX LoRA tensor name to GGUF LoRA tensor name.
// Input:  "model.layers.0.self_attn.q_proj.lora_a"
// Output: "blk.0.attn_q.weight.lora_a"
func MLXTensorToGGUF(mlxName string) (string, error) {
	m := mlxLoraKeyRe.FindStringSubmatch(mlxName)
	if m == nil {
		return "", fmt.Errorf("unrecognised MLX LoRA key: %s", mlxName)
	}

	layerNum := m[1]
	module := m[2]
	loraSuffix := m[3]

	ggufModule, ok := gemma3ModuleMap[module]
	if !ok {
		return "", fmt.Errorf("unknown module %q in %s", module, mlxName)
	}

	return fmt.Sprintf("blk.%s.%s.weight.%s", layerNum, ggufModule, loraSuffix), nil
}

// SafetensorsDtypeToGGML maps safetensors dtype strings to GGML types.
func SafetensorsDtypeToGGML(dtype string) (uint32, error) {
	switch dtype {
	case "F32":
		return ggmlTypeF32, nil
	case "F16":
		return ggmlTypeF16, nil
	case "BF16":
		return ggmlTypeBF16, nil
	default:
		return 0, fmt.Errorf("unsupported dtype %q for GGUF", dtype)
	}
}

// ConvertMLXtoGGUFLoRA converts an MLX LoRA adapter to GGUF LoRA format.
func ConvertMLXtoGGUFLoRA(safetensorsPath, configPath, outputPath, architecture string) error {
	cfgData, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("read config: %w", err)
	}

	var mlxConfig struct {
		LoraParameters struct {
			Rank  int     `json:"rank"`
			Scale float64 `json:"scale"`
		} `json:"lora_parameters"`
	}
	if err := json.Unmarshal(cfgData, &mlxConfig); err != nil {
		return fmt.Errorf("parse config: %w", err)
	}

	rank := mlxConfig.LoraParameters.Rank
	if rank == 0 {
		rank = 8
	}
	scale := mlxConfig.LoraParameters.Scale
	if scale == 0 {
		scale = 20.0
	}
	loraAlpha := float32(math.Round(scale * float64(rank)))

	tensors, tensorData, err := ReadSafetensors(safetensorsPath)
	if err != nil {
		return fmt.Errorf("read safetensors: %w", err)
	}
	log.Printf("loaded %d tensors from %s", len(tensors), safetensorsPath)

	var ggufTensors []ggufTensor
	for mlxKey, info := range tensors {
		ggufName, err := MLXTensorToGGUF(mlxKey)
		if err != nil {
			return err
		}

		ggmlType, err := SafetensorsDtypeToGGML(info.Dtype)
		if err != nil {
			return fmt.Errorf("tensor %s: %w", mlxKey, err)
		}

		data := GetTensorData(info, tensorData)

		if len(info.Shape) == 2 {
			rows, cols := info.Shape[0], info.Shape[1]
			switch info.Dtype {
			case "F32":
				data = TransposeFloat32(data, rows, cols)
			case "F16":
				data = TransposeFloat16(data, rows, cols)
			case "BF16":
				data = TransposeBFloat16(data, rows, cols)
			}
			ggufTensors = append(ggufTensors, ggufTensor{
				name:  ggufName,
				dims:  []uint64{uint64(rows), uint64(cols)},
				dtype: ggmlType,
				data:  data,
			})
		} else {
			dims := make([]uint64, len(info.Shape))
			for i, s := range info.Shape {
				dims[i] = uint64(s)
			}
			ggufTensors = append(ggufTensors, ggufTensor{
				name:  ggufName,
				dims:  dims,
				dtype: ggmlType,
				data:  data,
			})
		}
	}

	sort.Slice(ggufTensors, func(i, j int) bool {
		return ggufTensors[i].name < ggufTensors[j].name
	})

	metadata := []ggufMetadata{
		{key: "general.type", valueType: ggufTypeString, value: "adapter"},
		{key: "general.architecture", valueType: ggufTypeString, value: architecture},
		{key: "adapter.type", valueType: ggufTypeString, value: "lora"},
		{key: "adapter.lora.alpha", valueType: ggufTypeFloat32, value: loraAlpha},
	}

	if err := writeGGUF(outputPath, metadata, ggufTensors); err != nil {
		return fmt.Errorf("write GGUF: %w", err)
	}

	log.Printf("wrote GGUF LoRA: %s (%d tensors, alpha=%.0f)", outputPath, len(ggufTensors), loraAlpha)
	return nil
}

// writeGGUF writes a GGUF v3 file.
func writeGGUF(path string, metadata []ggufMetadata, tensors []ggufTensor) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	w := &ggufWriter{f: f}

	w.writeUint32(ggufMagic)
	w.writeUint32(ggufVersion)
	w.writeUint64(uint64(len(tensors)))
	w.writeUint64(uint64(len(metadata)))

	for _, kv := range metadata {
		w.writeString(kv.key)
		w.writeUint32(kv.valueType)
		switch kv.valueType {
		case ggufTypeString:
			w.writeString(kv.value.(string))
		case ggufTypeUint32:
			w.writeUint32(kv.value.(uint32))
		case ggufTypeFloat32:
			w.writeFloat32(kv.value.(float32))
		}
	}

	dataOffset := uint64(0)
	for _, t := range tensors {
		w.writeString(t.name)
		w.writeUint32(uint32(len(t.dims)))
		for _, d := range t.dims {
			w.writeUint64(d)
		}
		w.writeUint32(t.dtype)
		w.writeUint64(dataOffset)

		dataOffset += uint64(len(t.data))
		if rem := dataOffset % ggufAlignment; rem != 0 {
			dataOffset += ggufAlignment - rem
		}
	}

	pos := w.pos
	if rem := pos % ggufAlignment; rem != 0 {
		pad := ggufAlignment - rem
		w.writeBytes(make([]byte, pad))
	}

	for _, t := range tensors {
		w.writeBytes(t.data)
		if rem := uint64(len(t.data)) % ggufAlignment; rem != 0 {
			w.writeBytes(make([]byte, ggufAlignment-rem))
		}
	}

	return w.err
}

// ggufWriter tracks position and accumulates errors.
type ggufWriter struct {
	f   *os.File
	pos uint64
	err error
}

func (w *ggufWriter) writeBytes(b []byte) {
	if w.err != nil {
		return
	}
	n, err := w.f.Write(b)
	w.pos += uint64(n)
	if err != nil {
		w.err = err
	}
}

func (w *ggufWriter) writeUint32(v uint32) {
	b := make([]byte, 4)
	binary.LittleEndian.PutUint32(b, v)
	w.writeBytes(b)
}

func (w *ggufWriter) writeUint64(v uint64) {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, v)
	w.writeBytes(b)
}

func (w *ggufWriter) writeFloat32(v float32) {
	w.writeUint32(math.Float32bits(v))
}

func (w *ggufWriter) writeString(s string) {
	w.writeUint64(uint64(len(s)))
	w.writeBytes([]byte(s))
}

// DetectArchFromConfig tries to infer the model architecture from adapter_config.json.
func DetectArchFromConfig(configPath string) string {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return "gemma3"
	}
	var cfg struct {
		LoraParameters struct {
			Rank int `json:"rank"`
		} `json:"lora_parameters"`
	}
	json.Unmarshal(data, &cfg)
	return "gemma3"
}

// ArchitectureGGUFMap maps model tags to GGUF architecture names.
var ArchitectureGGUFMap = map[string]string{
	"gemma-3-1b":  "gemma3",
	"gemma-3-4b":  "gemma3",
	"gemma-3-12b": "gemma3",
	"gemma-3-27b": "gemma3",
}

// ModelTagToGGUFArch returns the GGUF architecture for a model tag.
func ModelTagToGGUFArch(modelTag string) string {
	if arch, ok := ArchitectureGGUFMap[modelTag]; ok {
		return arch
	}
	return "gemma3"
}

// GGUFModelBlobPath returns the path to the GGUF model blob in Ollama's store.
func GGUFModelBlobPath(ollamaModelsDir, modelName string) (string, error) {
	parts := strings.SplitN(modelName, ":", 2)
	family := parts[0]
	tag := "latest"
	if len(parts) > 1 {
		tag = parts[1]
	}

	manifestPath := fmt.Sprintf("%s/manifests/registry.ollama.ai/library/%s/%s", ollamaModelsDir, family, tag)
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return "", fmt.Errorf("read manifest %s: %w", manifestPath, err)
	}

	var manifest struct {
		Layers []struct {
			MediaType string `json:"mediaType"`
			Digest    string `json:"digest"`
		} `json:"layers"`
	}
	if err := json.Unmarshal(data, &manifest); err != nil {
		return "", fmt.Errorf("parse manifest: %w", err)
	}

	for _, layer := range manifest.Layers {
		if layer.MediaType == "application/vnd.ollama.image.model" {
			blobName := strings.Replace(layer.Digest, ":", "-", 1)
			return fmt.Sprintf("%s/blobs/%s", ollamaModelsDir, blobName), nil
		}
	}

	return "", fmt.Errorf("no model layer found in manifest for %s", modelName)
}

// ParseLayerFromTensorName extracts the layer number from a GGUF tensor name.
func ParseLayerFromTensorName(name string) (int, error) {
	re := regexp.MustCompile(`blk\.(\d+)\.`)
	m := re.FindStringSubmatch(name)
	if m == nil {
		return 0, fmt.Errorf("no layer number in %s", name)
	}
	return strconv.Atoi(m[1])
}
