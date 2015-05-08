// SAUCE (Standard Architecture for Universal Comment Extensions) parser.
package sauce

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	ASCIISub  = '\x1a'
	SAUCEDate = "19700101"
)

const (
	DATA_TYPE_NONE uint8 = iota
	DATA_TYPE_CHARACTER
	DATA_TYPE_BITMAP
	DATA_TYPE_VECTOR
	DATA_TYPE_AUDIO
	DATA_TYPE_BINARYTEXT
	DATA_TYPE_XBIN
	DATA_TYPE_ARCHIVE
	DATA_TYPE_EXECUTABLE
)

var (
	SAUCEID       = [5]byte{'S', 'A', 'U', 'C', 'E'}
	SAUCEVersion  = [2]byte{0, 0}
	SAUCEDataType = map[uint8]string{
		DATA_TYPE_NONE:       "None",
		DATA_TYPE_CHARACTER:  "Character",
		DATA_TYPE_BITMAP:     "Bitmap",
		DATA_TYPE_VECTOR:     "Vector",
		DATA_TYPE_AUDIO:      "Audio",
		DATA_TYPE_BINARYTEXT: "BinaryText",
		DATA_TYPE_XBIN:       "XBin",
		DATA_TYPE_ARCHIVE:    "Archive",
		DATA_TYPE_EXECUTABLE: "Executable",
	}
	SAUCEFileType = map[uint8]map[uint8]string{
		DATA_TYPE_CHARACTER: map[uint8]string{
			0: "ASCII",
			1: "ANSi",
			2: "ANSiMation",
			3: "RIP script",
			4: "PCBoard",
			5: "Avatar",
			6: "HTML",
			7: "Source",
			8: "Tundradraw",
		},
		DATA_TYPE_BITMAP: map[uint8]string{
			0:  "GIF",
			1:  "PCX",
			2:  "LBM/FF",
			3:  "TGA",
			4:  "FLI",
			5:  "FLC",
			6:  "BMP",
			7:  "GL",
			8:  "DL",
			9:  "WPG",
			10: "PNG",
			11: "JPG",
			12: "MPG",
			13: "AVI",
		},
		DATA_TYPE_VECTOR: map[uint8]string{
			0: "DXF",
			1: "DWG",
			2: "WPG",
			3: "3DS",
		},
		DATA_TYPE_AUDIO: map[uint8]string{
			0:  "MOD",
			1:  "669",
			2:  "STM",
			3:  "S3M",
			4:  "MTM",
			5:  "FAR",
			6:  "ULT",
			7:  "AMF",
			8:  "DMF",
			9:  "OKT",
			10: "ROL",
			11: "CMF",
			12: "MID",
			13: "SADT",
			14: "VOC",
			15: "WAV",
			16: "SMP8",
			17: "SMP8S",
			18: "SMP16",
			19: "SMP16S",
			20: "PATCH8",
			21: "PATCH16",
			22: "XM",
			23: "HSC",
			24: "IT",
		},
		DATA_TYPE_ARCHIVE: map[uint8]string{
			0: "ZIP",
			1: "ARJ",
			2: "LZH",
			3: "ARC",
			4: "TAR",
			5: "ZOO",
			6: "RAR",
			7: "UC2",
			8: "PAK",
			9: "SQZ",
		},
	}
)

// SAUCE (Standard Architecture for Universal Comment Extensions) record.
type SAUCE struct {
	ID       [5]byte
	Version  [2]byte
	Title    string
	Author   string
	Group    string
	Date     time.Time
	FileSize uint32
	DataType uint8
	FileType uint8
	TInfo    [4]uint16
	Comments uint8
	TFlags   uint8
	TInfos   [22]byte
}

// New creates an empty SAUCE record.
func New() *SAUCE {
	return &SAUCE{
		ID:      SAUCEID,
		Version: SAUCEVersion,
	}
}

// Parse SAUCE record from a file.
func Parse(filename string) (r *SAUCE, err error) {
	var f *os.File
	var i os.FileInfo

	f, err = os.Open(filename)
	if err != nil {
		return
	}
	defer f.Close()

	i, err = f.Stat()
	if err != nil {
		return
	}
	if i.Size() < 129 {
		return nil, errors.New("File too short")
	}

	var n int
	_, err = f.Seek(-128, 2)
	if err != nil {
		return
	}
	b := make([]byte, 128)
	n, err = f.Read(b)
	if err != nil {
		return
	}
	if n != 128 {
		return nil, errors.New("Short read")
	}
	//if b[0] != ASCIISub {
	//	return nil, errors.New("SUB character not found")
	//}
	if !bytes.Equal(b[0:5], SAUCEID[:]) {
		return nil, errors.New("No SAUCE record")
	}

	r = New()
	r.Title = strings.TrimSpace(string(b[7:41]))
	r.Author = strings.TrimSpace(string(b[41:61]))
	r.Group = strings.TrimSpace(string(b[61:81]))
	log.Printf("date: %q\n", string(b[82:90]))
	r.Date = r.parseDate(string(b[82:90]))
	r.FileSize = binary.LittleEndian.Uint32(b[91:95])
	r.DataType = uint8(b[94])
	r.FileType = uint8(b[95])
	r.TInfo[0] = binary.LittleEndian.Uint16(b[96:98])
	r.TInfo[1] = binary.LittleEndian.Uint16(b[98:100])
	r.TInfo[2] = binary.LittleEndian.Uint16(b[100:102])
	r.TInfo[3] = binary.LittleEndian.Uint16(b[102:104])
	return r, nil
}

func (r *SAUCE) parseDate(s string) time.Time {
	y, _ := strconv.Atoi(s[:4])
	m, _ := strconv.Atoi(s[4:6])
	d, _ := strconv.Atoi(s[6:8])
	return time.Date(y, time.Month(m), d, 0, 0, 0, 0, time.UTC)
}

// Dump the contents of the SAUCE record to stdout.
func (r *SAUCE) Dump() {
	fmt.Printf("id......: %s\n", string(r.ID[:]))
	fmt.Printf("version.: %d%d\n", r.Version[0], r.Version[1])
	fmt.Printf("title...: %s\n", r.Title)
	fmt.Printf("author..: %s\n", r.Author)
	fmt.Printf("group...: %s\n", r.Group)
	fmt.Printf("date....: %s\n", r.Date)
	fmt.Printf("filesize: %d\n", r.FileSize)
	fmt.Printf("datatype: %d (%s)\n", r.DataType, r.DataTypeString())
	if SAUCEFileType[r.DataType] != nil {
		fmt.Printf("filetype: %d (%s)\n", r.FileType, r.FileTypeString())
	} else {
		fmt.Printf("filetype: %d\n", r.FileType)
	}
	fmt.Printf("tinfo...: %d, %d, %d, %d\n", r.TInfo[0], r.TInfo[1], r.TInfo[2], r.TInfo[3])
	switch r.DataType {
	case 1:
		switch r.FileType {
		case 0, 1, 2, 4, 5, 8:
			w := r.TInfo[0]
			h := r.TInfo[1]
			if w == 0 {
				w = 80
			}
			fmt.Printf("size....: %d x %d characters\n", w, h)
		case 3:
			fmt.Printf("size....: %d x %d pixels\n", r.TInfo[0], r.TInfo[1])
		}
	case 2:
		fmt.Printf("size....: %d x %d pixels\n", r.TInfo[0], r.TInfo[1])
	}
}

// DataTypeString returns the DataType as string.
func (r *SAUCE) DataTypeString() string {
	return SAUCEDataType[r.DataType]
}

// FileTypeString returns the FileType as string.
func (r *SAUCE) FileTypeString() string {
	return SAUCEFileType[r.DataType][r.FileType]
}
