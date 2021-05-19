// Copyright 2016-2019 Authors of Cilium
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package labels

import (
	"bytes"
	"crypto/sha512"
	"encoding/json"
	"fmt"
	"net"
	"sort"
	"strings"
)

const (
	// PathDelimiter is the delimiter used in the labels paths.
	PathDelimiter = "."

	// IDNameHost is the label used for the hostname ID.
	IDNameHost = "host"

	// IDNameRemoteNode is the label used to describe the
	// ReservedIdentityRemoteNode
	IDNameRemoteNode = "remote-node"

	// IDNameWorld is the label used for the world ID.
	IDNameWorld = "world"

	// IDNameCluster is the label used to identify an unspecified endpoint
	// inside the cluster
	IDNameCluster = "cluster"

	// IDNameHealth is the label used for the local cilium-health endpoint
	IDNameHealth = "health"

	// IDNameInit is the label used to identify any endpoint that has not
	// received any labels yet.
	IDNameInit = "init"

	// IDNameNone is the label used to identify no endpoint or other L3 entity.
	// It will never be assigned and this "label" is here for consistency with
	// other Entities.
	IDNameNone = "none"

	// IDNameUnmanaged is the label used to identify unmanaged endpoints
	IDNameUnmanaged = "unmanaged"

	// IDNameUnknown is the label used to to identify an endpoint with an
	// unknown identity.
	IDNameUnknown = "unknown"
)

var (
	// LabelHealth is the label used for health.
	LabelHealth = Labels{IDNameHealth: NewLabel(IDNameHealth, "", LabelSourceReserved)}

	// LabelHost is the label used for the host endpoint.
	LabelHost = Labels{IDNameHost: NewLabel(IDNameHost, "", LabelSourceReserved)}
)

const (
	// LabelSourceUnspec is a label with unspecified source
	LabelSourceUnspec = "unspec"

	// LabelSourceAny is a label that matches any source
	LabelSourceAny = "any"

	// LabelSourceAnyKeyPrefix is prefix of a "any" label
	LabelSourceAnyKeyPrefix = LabelSourceAny + "."

	// LabelSourceK8s is a label imported from Kubernetes
	LabelSourceK8s = "k8s"

	// LabelSourceMesos is a label imported from Mesos
	LabelSourceMesos = "mesos"

	// LabelSourceK8sKeyPrefix is prefix of a Kubernetes label
	LabelSourceK8sKeyPrefix = LabelSourceK8s + "."

	// LabelSourceContainer is a label imported from the container runtime
	LabelSourceContainer = "container"

	// LabelSourceReserved is the label source for reserved types.
	LabelSourceReserved = "reserved"

	// LabelSourceCIDR is the label source for generated CIDRs.
	LabelSourceCIDR = "cidr"

	// LabelSourceReservedKeyPrefix is the prefix of a reserved label
	LabelSourceReservedKeyPrefix = LabelSourceReserved + "."

	// LabelKeyFixedIdentity is the label that can be used to define a fixed
	// identity.
	LabelKeyFixedIdentity = "io.cilium.fixed-identity"

	// LabelSourceCiliumGenerated is for labels auto-generated by cilium without
	// user input
	LabelSourceCiliumGenerated = "cilium-generated"
)

// Label is the Cilium's representation of a container label.
type Label struct {
	Key   string `json:"key"`
	Value string `json:"value,omitempty"`
	// Source can be one of the above values (e.g.: LabelSourceContainer).
	//
	// +kubebuilder:validation:Optional
	Source string `json:"source"`
}

// Labels is a map of labels where the map's key is the same as the label's key.
type Labels map[string]Label

// GetPrintableModel turns the Labels into a sorted list of strings
// representing the labels, with CIDRs deduplicated (ie, only provide the most
// specific CIDR).
func (l Labels) GetPrintableModel() (res []string) {
	cidr := ""
	prefixLength := 0
	for _, v := range l {
		if v.Source == LabelSourceCIDR {
			vStr := strings.Replace(v.String(), "-", ":", -1)
			prefix := strings.Replace(v.Key, "-", ":", -1)
			_, ipnet, _ := net.ParseCIDR(prefix)
			ones, _ := ipnet.Mask.Size()
			if ones > prefixLength {
				cidr = vStr
				prefixLength = ones
			}
			continue
		}
		res = append(res, v.String())
	}
	if cidr != "" {
		res = append(res, cidr)
	}

	sort.Strings(res)
	return res
}

// String returns the map of labels as human readable string
func (l Labels) String() string {
	return strings.Join(l.GetPrintableModel(), ",")
}

// AppendPrefixInKey appends the given prefix to all the Key's of the map and the
// respective Labels' Key.
func (l Labels) AppendPrefixInKey(prefix string) Labels {
	newLabels := Labels{}
	for k, v := range l {
		newLabels[prefix+k] = Label{
			Key:    prefix + v.Key,
			Value:  v.Value,
			Source: v.Source,
		}
	}
	return newLabels
}

// Equals returns true if the two Labels contain the same set of labels.
func (l Labels) Equals(other Labels) bool {
	if len(l) != len(other) {
		return false
	}

	for k, lbl1 := range l {
		if lbl2, ok := other[k]; ok {
			if lbl1.Source == lbl2.Source && lbl1.Key == lbl2.Key && lbl1.Value == lbl2.Value {
				continue
			}
		}
		return false
	}
	return true
}

// GetFromSource returns all labels that are from the given source.
func (l Labels) GetFromSource(source string) Labels {
	lbls := Labels{}
	for k, v := range l {
		if v.Source == source {
			lbls[k] = v
		}
	}
	return lbls
}

// NewLabel returns a new label from the given key, value and source. If source is empty,
// the default value will be LabelSourceUnspec. If key starts with '$', the source
// will be overwritten with LabelSourceReserved. If key contains ':', the value
// before ':' will be used as source if given source is empty, otherwise the value before
// ':' will be deleted and unused.
func NewLabel(key string, value string, source string) Label {
	var src string
	src, key = parseSource(key, ':')
	if source == "" {
		if src == "" {
			source = LabelSourceUnspec
		} else {
			source = src
		}
	}
	if src == LabelSourceReserved && key == "" {
		key = value
		value = ""
	}

	return Label{
		Key:    key,
		Value:  value,
		Source: source,
	}
}

// Equals returns true if source, Key and Value are equal and false otherwise.
func (l *Label) Equals(b *Label) bool {
	if !l.IsAnySource() && l.Source != b.Source {
		return false
	}
	return l.Key == b.Key && l.Value == b.Value
}

// IsAnySource return if the label was set with source "any".
func (l *Label) IsAnySource() bool {
	return l.Source == LabelSourceAny
}

// IsReservedSource return if the label was set with source "Reserved".
func (l *Label) IsReservedSource() bool {
	return l.Source == LabelSourceReserved
}

// matches returns true if l matches the target
func (l *Label) matches(target *Label) bool {
	return l.Equals(target)
}

// String returns the string representation of Label in the for of Source:Key=Value or
// Source:Key if Value is empty.
func (l *Label) String() string {
	if len(l.Value) != 0 {
		return l.Source + ":" + l.Key + "=" + l.Value
	}
	return l.Source + ":" + l.Key
}

// IsValid returns true if Key != "".
func (l *Label) IsValid() bool {
	return l.Key != ""
}

// UnmarshalJSON TODO create better explanation about unmarshall with examples
func (l *Label) UnmarshalJSON(data []byte) error {
	decoder := json.NewDecoder(bytes.NewReader(data))

	if l == nil {
		return fmt.Errorf("cannot unmarhshal to nil pointer")
	}

	if len(data) == 0 {
		return fmt.Errorf("invalid Label: empty data")
	}

	var aux struct {
		Source string `json:"source"`
		Key    string `json:"key"`
		Value  string `json:"value,omitempty"`
	}

	err := decoder.Decode(&aux)
	if err != nil {
		// If parsing of the full representation failed then try the short
		// form in the format:
		//
		// [SOURCE:]KEY[=VALUE]
		var aux string

		decoder = json.NewDecoder(bytes.NewReader(data))
		if err := decoder.Decode(&aux); err != nil {
			return fmt.Errorf("decode of Label as string failed: %+v", err)
		}

		if aux == "" {
			return fmt.Errorf("invalid Label: Failed to parse %s as a string", data)
		}

		*l = ParseLabel(aux)
	} else {
		if aux.Key == "" {
			return fmt.Errorf("invalid Label: '%s' does not contain label key", data)
		}

		l.Source = aux.Source
		l.Key = aux.Key
		l.Value = aux.Value
	}

	return nil
}

// GetExtendedKey returns the key of a label with the source encoded.
func (l *Label) GetExtendedKey() string {
	return l.Source + PathDelimiter + l.Key
}

// GetCiliumKeyFrom returns the label's source and key from the an extended key
// in the format SOURCE:KEY.
func GetCiliumKeyFrom(extKey string) string {
	i := strings.IndexByte(extKey, PathDelimiter[0])
	if i >= 0 {
		return extKey[:i] + ":" + extKey[i+1:]
	}
	return LabelSourceAny + ":" + extKey
}

// GetExtendedKeyFrom returns the extended key of a label string.
// For example:
// `k8s:foo=bar` returns `k8s.foo`
// `container:foo=bar` returns `container.foo`
// `foo=bar` returns `any.foo=bar`
func GetExtendedKeyFrom(str string) string {
	src, next := parseSource(str, ':')
	if src == "" {
		src = LabelSourceAny
	}
	// Remove an eventually value
	i := strings.IndexByte(next, '=')
	if i >= 0 {
		return src + PathDelimiter + next[:i]
	}
	return src + PathDelimiter + next
}

// Map2Labels transforms in the form: map[key(string)]value(string) into Labels. The
// source argument will overwrite the source written in the key of the given map.
// Example:
// l := Map2Labels(map[string]string{"k8s:foo": "bar"}, "cilium")
// fmt.Printf("%+v\n", l)
//   map[string]Label{"foo":Label{Key:"foo", Value:"bar", Source:"cilium"}}
func Map2Labels(m map[string]string, source string) Labels {
	o := Labels{}
	for k, v := range m {
		l := NewLabel(k, v, source)
		o[l.Key] = l
	}
	return o
}

// StringMap converts Labels into map[string]string
func (l Labels) StringMap() map[string]string {
	o := map[string]string{}
	for _, v := range l {
		o[v.Source+":"+v.Key] = v.Value
	}
	return o
}

// StringMap converts Labels into map[string]string
func (l Labels) K8sStringMap() map[string]string {
	o := map[string]string{}
	for _, v := range l {
		if v.Source == LabelSourceK8s || v.Source == LabelSourceAny || v.Source == LabelSourceUnspec {
			o[v.Key] = v.Value
		} else {
			o[v.Source+"."+v.Key] = v.Value
		}
	}
	return o
}

// NewLabelsFromModel creates labels from string array.
func NewLabelsFromModel(base []string) Labels {
	lbls := make(Labels, len(base))
	for _, v := range base {
		if lbl := ParseLabel(v); lbl.Key != "" {
			lbls[lbl.Key] = lbl
		}
	}

	return lbls
}

// NewLabelsFromSortedList returns labels based on the output of SortedList()
func NewLabelsFromSortedList(list string) Labels {
	return NewLabelsFromModel(strings.Split(list, ";"))
}

// NewSelectLabelArrayFromModel parses a slice of strings and converts them
// into an array of selecting labels, sorted by the key.
func NewSelectLabelArrayFromModel(base []string) LabelArray {
	lbls := make(LabelArray, 0, len(base))
	for i := range base {
		lbls = append(lbls, ParseSelectLabel(base[i]))
	}

	return lbls.Sort()
}

// GetModel returns model with all the values of the labels.
func (l Labels) GetModel() []string {
	res := make([]string, 0, len(l))
	for _, v := range l {
		res = append(res, v.String())
	}
	return res
}

// MergeLabels merges labels from into to. It overwrites all labels with the same Key as
// from written into to.
// Example:
// to := Labels{Label{key1, value1, source1}, Label{key2, value3, source4}}
// from := Labels{Label{key1, value3, source4}}
// to.MergeLabels(from)
// fmt.Printf("%+v\n", to)
//   Labels{Label{key1, value3, source4}, Label{key2, value3, source4}}
func (l Labels) MergeLabels(from Labels) {
	for k, v := range from {
		l[k] = v
	}
}

// SHA256Sum calculates l' internal SHA256Sum. For a particular set of labels is
// guarantee that it will always have the same SHA256Sum.
func (l Labels) SHA256Sum() string {
	return fmt.Sprintf("%x", sha512.Sum512_256(l.SortedList()))
}

// FormatForKVStore returns the label as a formatted string, ending in
// a semicolon
//
// DO NOT BREAK THE FORMAT OF THIS. THE RETURNED STRING IS USED AS
// PART OF THE KEY IN THE KEY-VALUE STORE.
//
// Non-pointer receiver allows this to be called on a value in a map.
func (l Label) FormatForKVStore() string {
	// We don't care if the values already have a '=' since this method is
	// only used to calculate a SHA256Sum
	//
	// We absolutely care that the final character is a semi-colon.
	// Identity allocation in the kvstore depends on this (see
	// kvstore.prefixMatchesKey())
	return fmt.Sprintf(`%s:%s=%s;`, l.Source, l.Key, l.Value)
}

// SortedList returns the labels as a sorted list, separated by semicolon
//
// DO NOT BREAK THE FORMAT OF THIS. THE RETURNED STRING IS USED AS KEY IN
// THE KEY-VALUE STORE.
func (l Labels) SortedList() []byte {
	keys := make([]string, 0, len(l))
	for k := range l {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	result := ""
	for _, k := range keys {
		result += l[k].FormatForKVStore()
	}

	return []byte(result)
}

// ToSlice returns a slice of label with the values of the given
// Labels' map, sorted by the key.
func (l Labels) ToSlice() []Label {
	return l.LabelArray()
}

// LabelArray returns the labels as label array, sorted by the key.
func (l Labels) LabelArray() LabelArray {
	labels := make(LabelArray, 0, len(l))
	for _, v := range l {
		labels = append(labels, v)
	}
	return labels.Sort()
}

// FindReserved locates all labels with reserved source in the labels and
// returns a copy of them. If there are no reserved labels, returns nil.
// TODO: return LabelArray as it is likely faster
func (l Labels) FindReserved() Labels {
	lbls := Labels{}

	for k, lbl := range l {
		if lbl.Source == LabelSourceReserved {
			lbls[k] = lbl
		}
	}

	if len(lbls) > 0 {
		return lbls
	}
	return nil
}

// IsReserved returns true if any of the labels has a reserved source.
func (l Labels) IsReserved() bool {
	for _, lbl := range l {
		if lbl.Source == LabelSourceReserved {
			return true
		}
	}
	return false
}

// parseSource returns the parsed source of the given str. It also returns the next piece
// of text that is after the source.
// Example:
//  src, next := parseSource("foo:bar==value")
// Println(src) // foo
// Println(next) // bar==value
// For Cilium format 'delim' must be passed in as ':'
// For k8s format 'delim' must be passed in as '.'
func parseSource(str string, delim byte) (src, next string) {
	if str == "" {
		return "", ""
	}
	if str[0] == '$' {
		return LabelSourceReserved, str[1:]
	}
	i := strings.IndexByte(str, delim)
	if i < 0 {
		if delim != '.' && strings.HasPrefix(str, LabelSourceReservedKeyPrefix) {
			return LabelSourceReserved, strings.TrimPrefix(str, LabelSourceReservedKeyPrefix)
		}
		return "", str
	}
	return str[:i], str[i+1:]
}

// ParseLabel returns the label representation of the given string. The str should be
// in the form of Source:Key=Value or Source:Key if Value is empty. It also parses short
// forms, for example: $host will be Label{Key: "host", Source: "reserved", Value: ""}.
func ParseLabel(str string) Label {
	return parseLabel(str, ':')
}

// parseLabel returns the label representation of the given string by value.
// For Cilium format 'delim' must be passed in as ':'
// For k8s format 'delim' must be passed in as '.'
func parseLabel(str string, delim byte) (lbl Label) {
	src, next := parseSource(str, delim)
	if src != "" {
		lbl.Source = src
	} else {
		lbl.Source = LabelSourceUnspec
	}

	i := strings.IndexByte(next, '=')
	if i < 0 {
		lbl.Key = next
	} else {
		if i == 0 && src == LabelSourceReserved {
			lbl.Key = next[i+1:]
		} else {
			lbl.Key = next[:i]
			lbl.Value = next[i+1:]
		}
	}
	return lbl
}

// ParseSelectLabel returns a selecting label representation of the given
// string. Unlike ParseLabel, if source is unspecified, the source defaults to
// LabelSourceAny
func ParseSelectLabel(str string) Label {
	return parseSelectLabel(str, ':')
}

// parseSelectLabel returns a selecting label representation of the given
// string by value.
// For Cilium format 'delim' must be passed in as ':'
// For k8s format 'delim' must be passed in as '.'
func parseSelectLabel(str string, delim byte) Label {
	lbl := parseLabel(str, delim)

	if lbl.Source == LabelSourceUnspec {
		lbl.Source = LabelSourceAny
	}

	return lbl
}

// generateLabelString generates the string representation of a label with
// the provided source, key, and value in the format "source:key=value".
func generateLabelString(source, key, value string) string {
	return fmt.Sprintf("%s:%s=%s", source, key, value)
}

// GenerateK8sLabelString generates the string representation of a label with
// the provided source, key, and value in the format "LabelSourceK8s:key=value".
func GenerateK8sLabelString(k, v string) string {
	return generateLabelString(LabelSourceK8s, k, v)
}