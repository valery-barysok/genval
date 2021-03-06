package types

import (
	"fmt"
	"io"
)

func NewMap(key, value TypeDef) *typeMap {
	return &typeMap{key: key, value: value}
}

type typeMap struct {
	min   *string
	max   *string
	key   TypeDef
	value TypeDef
}

func (t typeMap) Type() string {
	return "map"
}

func (t *typeMap) SetTag(tag Tag) error {
	switch tag.Key() {
	case MapMinItemsKey:
		st := tag.(SimpleTag)
		t.min = &st.Param
	case MapMaxItemsKey:
		st := tag.(SimpleTag)
		t.max = &st.Param
	case MapKeyKey:
		scope := tag.(ScopeTag)
		for _, it := range scope.InnerTags {
			if err := t.key.SetTag(it); err != nil {
				return fmt.Errorf("set item tags for key failed, tag %+v, err %s", it, err)
			}
		}
	case MapValueKey:
		scope := tag.(ScopeTag)
		for _, it := range scope.InnerTags {
			if err := t.value.SetTag(it); err != nil {
				return fmt.Errorf("set item tags for value failed, tag %+v, err %s", it, err)
			}
		}
	default:
		return ErrUnusedTag
	}
	return nil
}

func (t typeMap) Generate(w io.Writer, cfg GenConfig, name Name) {
	if t.min != nil {
		if *t.min != "0" {
			cfg.AddImport("fmt")
			fmt.Fprintf(w, "if len(%s) < %s {\n", name.Full(), *t.min)
			fmt.Fprintf(w, "    return fmt.Errorf(\"map %s has less items than %s \" )\n", name.FieldName(), *t.min)
			fmt.Fprintf(w, "}\n")
		}
	}
	if t.max != nil {
		cfg.AddImport("fmt")
		fmt.Fprintf(w, "if len(%s) > %s {\n", name.Full(), *t.max)
		fmt.Fprintf(w, "    return fmt.Errorf(\"map %s has more items than %s \" )\n", name.FieldName(), *t.max)
		fmt.Fprintf(w, "}\n")
	}
	fmt.Fprintf(w, "for k, v := range %s {\n", name.Full())
	fmt.Fprintf(w, "	_ = k \n")
	fmt.Fprintf(w, "	_ = v \n")
	t.key.Generate(w, cfg, NewSimpleName("k"))
	t.value.Generate(w, cfg, NewSimpleName("v"))
	fmt.Fprintf(w, "}\n")
}

func (t typeMap) Validate() error {
	if err := validateMinMax(
		t.min,
		t.max,
		func(min float64) error {
			if min < 0 {
				return fmt.Errorf("min map items can't be less than 0: %f", min)
			}
			return nil
		},
		func(max float64) error {
			if max < 0 {
				return fmt.Errorf("max map items can't be less than 0: %f", max)
			}
			return nil
		},
	); err != nil {
		return err
	}
	if err := t.key.Validate(); err != nil {
		return err
	}
	return t.value.Validate()
}
