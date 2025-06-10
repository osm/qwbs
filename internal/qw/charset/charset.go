package charset

import "strings"

func Parse(input string) string {
	var result strings.Builder

	special := [32]string{
		"", "", "", "", "", ".", "", "",
		"", "", "", "", "", ">", ".", ".",
		"[", "]", "0", "1", "2", "3", "4", "5",
		"6", "7", "8", "9", ".", "<", "=", ">",
	}

	for _, b := range []byte(input) {
		b &= 0x7f

		if b == 0x7f {
			continue
		}

		if b < 32 {
			result.WriteString(special[b])
		} else {
			result.WriteByte(b)
		}
	}

	return result.String()
}
