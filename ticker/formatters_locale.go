package ticker

import (
	"os"
	"strconv"
	"strings"

	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

func scalarLocaleLanguage() language.Tag {
	for _, key := range []string{"LC_ALL", "LC_NUMERIC", "LANG"} {
		value := strings.TrimSpace(os.Getenv(key))
		if value == "" {
			continue
		}
		value = strings.SplitN(value, ".", 2)[0]
		value = strings.SplitN(value, "@", 2)[0]
		if value == "C" || value == "POSIX" {
			return language.English
		}
		return language.Make(strings.ReplaceAll(value, "_", "-"))
	}
	return language.English
}

func scalarFormatFixed(f ScalarFormatter, value float64, precision int) string {
	if !f.UseLocale {
		return strconv.FormatFloat(value, 'f', precision, 64)
	}
	return message.NewPrinter(scalarLocaleLanguage()).Sprintf("%.*f", precision, value)
}

func scalarFormatGeneral(f ScalarFormatter, value float64, precision int) string {
	if !f.UseLocale {
		return strconv.FormatFloat(value, 'g', precision, 64)
	}
	return message.NewPrinter(scalarLocaleLanguage()).Sprintf("%.*g", precision, value)
}

func scalarFormatInteger(f ScalarFormatter, value int) string {
	if !f.UseLocale {
		return strconv.Itoa(value)
	}
	return message.NewPrinter(scalarLocaleLanguage()).Sprintf("%d", value)
}

func scalarTrimFixed(value string, precision int) string {
	if precision <= 0 {
		return value
	}
	value = strings.TrimRight(value, "0")
	value = strings.TrimSuffix(value, ".")
	return strings.TrimSuffix(value, ",")
}

func scalarMathTextNumber(f ScalarFormatter, value string) string {
	if !f.UseLocale {
		return value
	}
	return strings.ReplaceAll(value, ",", "{,}")
}

func scalarMathTextLabel(f ScalarFormatter, value string) string {
	return `$\mathdefault{` + scalarMathTextNumber(f, value) + `}$`
}
