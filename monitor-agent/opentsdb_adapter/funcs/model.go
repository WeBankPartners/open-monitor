package funcs

import (
	"unicode/utf8"
	"bytes"
	"fmt"
	"github.com/gogo/protobuf/proto"
	"time"
)

type StoreSamplesRequest struct {
	Metric    TagValue            `json:"metric"`
	Timestamp int64               `json:"timestamp"`
	Value     float64             `json:"value"`
	Tags      map[string]TagValue `json:"tags"`
}

type TagValue LabelValue

func (tv TagValue) MarshalJSON() ([]byte, error) {
	length := len(tv)
	// Need at least two more bytes than in tv.
	result := bytes.NewBuffer(make([]byte, 0, length+2))
	result.WriteByte('"')
	for i := 0; i < length; i++ {
		b := tv[i]
		switch {
		case (b >= '-' && b <= '9') || // '-', '.', '/', 0-9
			(b >= 'A' && b <= 'Z') ||
			(b >= 'a' && b <= 'z'):
			result.WriteByte(b)
		case b == '_':
			result.WriteString("__")
		case b == ':':
			result.WriteString("_.")
		default:
			result.WriteString(fmt.Sprintf("_%X", b))
		}
	}
	result.WriteByte('"')
	return result.Bytes(), nil
}

func (tv *TagValue) UnmarshalJSON(json []byte) error {
	escapeLevel := 0 // How many bytes after '_'.
	var parsedByte byte

	// Might need fewer bytes, but let's avoid realloc.
	result := bytes.NewBuffer(make([]byte, 0, len(json)-2))

	for i, b := range json {
		if i == 0 {
			if b != '"' {
				return errorF("expected '\"', got %q", b)
			}
			continue
		}
		if i == len(json)-1 {
			if b != '"' {
				return errorF("expected '\"', got %q", b)
			}
			break
		}
		switch escapeLevel {
		case 0:
			if b == '_' {
				escapeLevel = 1
				continue
			}
			result.WriteByte(b)
		case 1:
			switch {
			case b == '_':
				result.WriteByte('_')
				escapeLevel = 0
			case b == '.':
				result.WriteByte(':')
				escapeLevel = 0
			case b >= '0' && b <= '9':
				parsedByte = (b - 48) << 4
				escapeLevel = 2
			case b >= 'A' && b <= 'F': // A-F
				parsedByte = (b - 55) << 4
				escapeLevel = 2
			default:
				return errorF(
					"illegal escape sequence at byte %d (%c)",
					i, b,
				)
			}
		case 2:
			switch {
			case b >= '0' && b <= '9':
				parsedByte += b - 48
			case b >= 'A' && b <= 'F': // A-F
				parsedByte += b - 55
			default:
				return errorF(
					"illegal escape sequence at byte %d (%c)",
					i, b,
				)
			}
			result.WriteByte(parsedByte)
			escapeLevel = 0
		default:
			panic("unexpected escape level")
		}
	}
	*tv = TagValue(result.String())
	return nil
}

// A LabelValue is an associated value for a LabelName.
type LabelName string
type LabelValue string

// IsValid returns true iff the string is a valid UTF8.
func (lv LabelValue) IsValid() bool {
	return utf8.ValidString(string(lv))
}

func errorF(format string, args ...interface{}) error {
	return fmt.Errorf(format, args...)
}

type WriteRequest struct {
	Timeseries           []TimeSeries `protobuf:"bytes,1,rep,name=timeseries,proto3" json:"timeseries"`
}

func (m *WriteRequest) Reset()         { *m = WriteRequest{} }
func (m *WriteRequest) String() string { return proto.CompactTextString(m) }
func (*WriteRequest) ProtoMessage()    {}
func (*WriteRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_eefc82927d57d89b, []int{0}
}

func (m *WriteRequest) GetTimeseries() []TimeSeries {
	if m != nil {
		return m.Timeseries
	}
	return nil
}

type TimeSeries struct {
	Labels               []Label  `protobuf:"bytes,1,rep,name=labels,proto3" json:"labels"`
	Samples              []Sample `protobuf:"bytes,2,rep,name=samples,proto3" json:"samples"`
}

type Label struct {
	Name                 string   `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	Value                string   `protobuf:"bytes,2,opt,name=value,proto3" json:"value,omitempty"`
}

type Sample struct {
	Value                float64  `protobuf:"fixed64,1,opt,name=value,proto3" json:"value,omitempty"`
	Timestamp            int64    `protobuf:"varint,2,opt,name=timestamp,proto3" json:"timestamp,omitempty"`
}

var fileDescriptor_eefc82927d57d89b = []byte{
	// 466 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x8c, 0x93, 0xcf, 0x6e, 0xd3, 0x40,
	0x10, 0xc6, 0xbb, 0x4d, 0xdb, 0xa0, 0x71, 0x88, 0xc2, 0xb6, 0x25, 0xa6, 0x87, 0x34, 0xb2, 0x38,
	0x58, 0x2a, 0x0a, 0x22, 0x54, 0x9c, 0x38, 0x90, 0x96, 0x48, 0x45, 0x24, 0xfc, 0x59, 0x07, 0x81,
	0x10, 0x92, 0xe5, 0xd8, 0xa3, 0xc6, 0xa2, 0xfe, 0xd3, 0xdd, 0xb5, 0xd4, 0xbc, 0x1e, 0xa7, 0x9e,
	0x10, 0x4f, 0x80, 0x50, 0x9e, 0x04, 0xed, 0xda, 0x0e, 0x1b, 0xb8, 0x70, 0x5b, 0x7f, 0xdf, 0x37,
	0x3f, 0xef, 0x8c, 0xc7, 0xd0, 0xe2, 0x98, 0x64, 0x12, 0x07, 0x39, 0xcf, 0x64, 0x46, 0x21, 0xe7,
	0x59, 0x82, 0x72, 0x81, 0x85, 0x38, 0xb2, 0xe4, 0x32, 0x47, 0x51, 0x1a, 0x47, 0x07, 0x97, 0xd9,
	0x65, 0xa6, 0x8f, 0x8f, 0xd5, 0xa9, 0x54, 0x9d, 0x09, 0xb4, 0x3e, 0xf2, 0x58, 0x22, 0xc3, 0xeb,
	0x02, 0x85, 0xa4, 0xcf, 0x01, 0x64, 0x9c, 0xa0, 0x40, 0x1e, 0xa3, 0xb0, 0x49, 0xbf, 0xe1, 0x5a,
	0xc3, 0xfb, 0x83, 0x3f, 0xcc, 0xc1, 0x2c, 0x4e, 0xd0, 0xd3, 0xee, 0xd9, 0xce, 0xed, 0xcf, 0xe3,
	0x2d, 0x66, 0xe4, 0x9d, 0xef, 0x04, 0x2c, 0x86, 0x41, 0x54, 0xd3, 0x4e, 0xa0, 0x79, 0x5d, 0x98,
	0xa8, 0x7b, 0x26, 0xea, 0x7d, 0x81, 0x7c, 0xc9, 0xea, 0x04, 0xfd, 0x02, 0xdd, 0x20, 0x0c, 0x31,
	0x97, 0x18, 0xf9, 0x1c, 0x45, 0x9e, 0xa5, 0x02, 0x7d, 0xdd, 0x81, 0xbd, 0xdd, 0x6f, 0xb8, 0xed,
	0xe1, 0x43, 0xb3, 0xd8, 0x78, 0xcd, 0x80, 0x55, 0xe9, 0xd9, 0x32, 0x47, 0x76, 0x58, 0x43, 0x4c,
	0x55, 0x38, 0xa7, 0xd0, 0x32, 0x05, 0x6a, 0x41, 0xd3, 0x1b, 0x4d, 0xdf, 0x4d, 0xc6, 0x5e, 0x67,
	0x8b, 0x76, 0x61, 0xdf, 0x9b, 0xb1, 0xf1, 0x68, 0x3a, 0x7e, 0xe9, 0x7f, 0x7a, 0xcb, 0xfc, 0xf3,
	0x8b, 0x0f, 0x6f, 0x5e, 0x7b, 0x1d, 0xe2, 0x8c, 0x54, 0x55, 0xb0, 0x46, 0xd1, 0x27, 0xd0, 0xe4,
	0x28, 0x8a, 0x2b, 0x59, 0x37, 0xd4, 0xfd, 0xb7, 0x21, 0xed, 0xb3, 0x3a, 0xe7, 0x7c, 0x23, 0xb0,
	0xab, 0x0d, 0xfa, 0x08, 0xa8, 0x90, 0x01, 0x97, 0xbe, 0x9e, 0x98, 0x0c, 0x92, 0xdc, 0x4f, 0x14,
	0x87, 0xb8, 0x0d, 0xd6, 0xd1, 0xce, 0xac, 0x36, 0xa6, 0x82, 0xba, 0xd0, 0xc1, 0x34, 0xda, 0xcc,
	0x6e, 0xeb, 0x6c, 0x1b, 0xd3, 0xc8, 0x4c, 0x9e, 0xc2, 0x9d, 0x24, 0x90, 0xe1, 0x02, 0xb9, 0xb0,
	0x1b, 0xfa, 0x56, 0xb6, 0x79, 0xab, 0x49, 0x30, 0xc7, 0xab, 0x69, 0x19, 0x60, 0xeb, 0x24, 0x3d,
	0x81, 0xdd, 0x45, 0x9c, 0x4a, 0x61, 0xef, 0xf4, 0x89, 0x6b, 0x0d, 0x0f, 0xff, 0x1e, 0xee, 0x85,
	0x32, 0x59, 0x99, 0x71, 0xc6, 0x60, 0x19, 0xcd, 0xd1, 0x67, 0xff, 0xbf, 0x25, 0x1b, 0xfb, 0x71,
	0x03, 0xfb, 0xe7, 0x8b, 0x22, 0xfd, 0xaa, 0x3e, 0x8e, 0x31, 0xd5, 0x17, 0xd0, 0x0e, 0x4b, 0xd9,
	0xdf, 0x40, 0x3e, 0x30, 0x91, 0x55, 0x61, 0x45, 0xbd, 0x1b, 0x9a, 0x8f, 0xf4, 0x18, 0x2c, 0xb5,
	0x46, 0x4b, 0x3f, 0x4e, 0x23, 0xbc, 0xa9, 0xe6, 0x04, 0x5a, 0x7a, 0xa5, 0x94, 0xb3, 0x83, 0xdb,
	0x55, 0x8f, 0xfc, 0x58, 0xf5, 0xc8, 0xaf, 0x55, 0x8f, 0x7c, 0xde, 0x53, 0xdc, 0x7c, 0x3e, 0xdf,
	0xd3, 0x3f, 0xc1, 0xd3, 0xdf, 0x01, 0x00, 0x00, 0xff, 0xff, 0x9a, 0xb6, 0x6b, 0xcd, 0x43, 0x03,
	0x00, 0x00,
}

type Samples []*PrometheusSample

type PrometheusSample struct {
	Metric    Metric      `json:"metric"`
	Value     SampleValue `json:"value"`
	Timestamp Time        `json:"timestamp"`
}

type SampleValue float64

type Time int64

func (t Time) Unix() int64 {
	return int64(t) / int64(time.Second / time.Millisecond)
}

type Metric LabelSet

type LabelSet map[LabelName]LabelValue

type OpenTsdbQuery struct {
	Start   int64                    `json:"start"`
	End     int64                    `json:"end"`
	Queries []map[string]interface{} `json:"queries"`
}

type OpenTsdbResponse struct {
	Metric     string             `json:"metric"`
	Tags       map[string]string  `json:"tags"`
	DataPoints map[string]float64 `json:"dps"`
}

type SerialModel struct {
	Type  string  `json:"type"`
	Name  string  `json:"name"`
	Data  [][]float64  `json:"data"`
}

type QueryMonitorData struct{
	Start  int64  `json:"start"`
	End    int64  `json:"end"`
	Endpoint  []string  `json:"endpoint"`
	Metric  []string  `json:"metric"`
	ComputeRate  bool  `json:"compute_rate"`
}

type QueryResponseDto struct {
	Code  int  `json:"code"`
	Message  string  `json:"message"`
	Data  []*SerialModel  `json:"data"`
}

type DataSort [][]float64

func (s DataSort) Len() int {
	return len(s)
}

func (s DataSort) Swap(i,j int)  {
	s[i], s[j] = s[j], s[i]
}

func (s DataSort) Less(i,j int) bool {
	return s[i][0] < s[j][0]
}