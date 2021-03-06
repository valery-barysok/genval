package types

import (
	"fmt"
	"io"
)

func NewArray(inner TypeDef) *typeArray {
	return &typeArray{innerType: inner}
}

type typeArray struct {
	min       *string
	max       *string
	innerType TypeDef
}

func (t typeArray) Type() string {
	return t.innerType.Type()
}

func (t *typeArray) SetTag(tag Tag) error {
	switch tag.Key() {
	case ArrayMinItemsKey:
		st := tag.(SimpleTag)
		t.min = &st.Param
	case ArrayMaxItemsKey:
		st := tag.(SimpleTag)
		t.max = &st.Param
	case ArrayItemKey:
		scope := tag.(ScopeTag)
		for _, it := range scope.InnerTags {
			if err := t.innerType.SetTag(it); err != nil {
				return fmt.Errorf("set item tags failed for %+v, err: %s", it, err)
			}
		}
	default:
		return ErrUnusedTag
	}
	return nil
}

func (t typeArray) Generate(w io.Writer, cfg GenConfig, name Name) {
	if t.min != nil {
		if *t.min != "0" {
			cfg.AddImport("fmt")
			fmt.Fprintf(w, "if len(%s) < %s {\n", name.Full(), *t.min)
			fmt.Fprintf(w, "    return fmt.Errorf(\"array %s has less items than %s \" )\n", name.FieldName(), *t.min)
			fmt.Fprintf(w, "}\n")
		}
	}
	if t.max != nil {
		cfg.AddImport("fmt")
		fmt.Fprintf(w, "if len(%s) > %s {\n", name.Full(), *t.max)
		fmt.Fprintf(w, "    return fmt.Errorf(\"array %s has more items than %s \" )\n", name.FieldName(), *t.max)
		fmt.Fprintf(w, "}\n")
	}
	fmt.Fprintf(w, "for _, x := range %s {\n", name.Full())
	fmt.Fprintf(w, "	_ = x \n")
	t.innerType.Generate(w, cfg, NewSimpleName("x"))
	fmt.Fprintf(w, "}\n")
}

func (t typeArray) Validate() error {
	if err := validateMinMax(
		t.min,
		t.max,
		func(min float64) error {
			if min < 0 {
				return fmt.Errorf("min items can't be less than 0: %f", min)
			}
			return nil
		},
		func(max float64) error {
			if max < 0 {
				return fmt.Errorf("max items can't be less than 0: %f", max)
			}
			return nil
		},
	); err != nil {
		return err
	}
	return t.innerType.Validate()
}
