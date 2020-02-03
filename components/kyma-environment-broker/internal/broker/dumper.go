package broker

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/pivotal-cf/brokerapi/domain"
	"github.com/sanity-io/litter"
)

// Dumper is the simplest implementation just for dumping also RawContext and RawParameters
// as strings instead of byte array.
type Dumper struct {
	underlying *litter.Options
}

// NewDumper returns new instance of the Dumper
func NewDumper() (*Dumper, error) {
	fieldExclusions, err := regexp.Compile(`(^(XXX_.*)$|RawContext|RawParameters)`)
	if err != nil {
		return nil, fmt.Errorf("while compiling regex: %v", err)
	}
	return &Dumper{
		underlying: &litter.Options{
			StripPackageNames: false,
			HidePrivateFields: true,
			FieldExclusions:   fieldExclusions,
			Separator:         " ",
		},
	}, nil
}

// Dump a value to stdout according to the options.
// Understands the Details struct from the brokerapi domain
func (d *Dumper) Dump(value ...interface{}) {
	b := strings.Builder{}
	b.WriteString(d.underlying.Sdump(value...))

	for _, v := range value {
		switch d := v.(type) {
		case domain.ProvisionDetails:
			b.WriteString(fmt.Sprintf("\nRawContext: %s", d.RawContext))
			b.WriteString(fmt.Sprintf("\nRawParameters: %s", d.RawParameters))
		case domain.UpdateDetails:
			b.WriteString(fmt.Sprintf("\nRawContext: %s", d.RawContext))
			b.WriteString(fmt.Sprintf("\nRawParameters: %s", d.RawParameters))
		case domain.BindDetails:
			b.WriteString(fmt.Sprintf("\nRawContext: %s", d.RawContext))
			b.WriteString(fmt.Sprintf("\nRawParameters: %s", d.RawParameters))
		}
	}

	b.WriteString("\n")
	fmt.Print(b.String())
}

// DumyDumper does nothing. Suitable for testing to silence the output.
type DumyDumper struct{}

func (d *DumyDumper) Dump(value ...interface{}) {}
