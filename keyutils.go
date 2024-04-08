package natskv

import (
	"github.com/cristalhq/base64"
	"strings"
)

type keyUtils struct {
	options *Config
}

func (s *keyUtils) isInDirectory(directory, key string) bool {
	keyParts := strings.Split(key, ".")
	dirParts := strings.Split(directory, ".")

	if len(keyParts) > 0 && len(dirParts) == 1 {
		isExists := strings.HasPrefix(key, directory+".")
		return isExists
	}
	return strings.HasPrefix(key, directory)
}

func (s *keyUtils) normalizeKey(key string) string {
	var parts []string
	for _, part := range strings.Split(key, "/") {
		if part == "" {
			continue
		}
		if s.options != nil && s.options.EncodeKey {
			parts = append(parts, base64.StdEncoding.EncodeToString([]byte(part)))
		} else {
			parts = append(parts, part)
		}
	}
	return strings.Join(parts, ".")
}

func (s *keyUtils) decodeKey(key string) string {
	var parts []string
	for _, part := range strings.Split(key, ".") {
		if s.options != nil && s.options.EncodeKey {
			d, err := base64.StdEncoding.DecodeToString([]byte(part))
			if err != nil {
				parts = append(parts, part)
			} else {
				parts = append(parts, d)
			}
		} else {
			parts = append(parts, part)
		}
	}
	return strings.Join(parts, "/")
}
