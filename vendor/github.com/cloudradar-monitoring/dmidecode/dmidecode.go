// +build !windows

package dmidecode

import (
	"fmt"
	"io"
	"io/ioutil"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var handleRegex *regexp.Regexp
var inBlockRegex *regexp.Regexp
var recordRegex *regexp.Regexp
var recordRegex2 *regexp.Regexp
var sizeTypeRegex *regexp.Regexp

type typeDecoders map[ReqType]func() interface{}

var decoders = typeDecoders{
	TypeBIOSInfo:                      newReqBiosInfo,
	TypeSystem:                        newReqSystem,
	TypeBaseBoard:                     newReqBaseBoard,
	TypeChassis:                       newReqChassis,
	TypeProcessor:                     newReqProcessor,
	TypeMemoryController:              newReqMemoryController,
	TypeMemoryModule:                  newReqMemoryModule,
	TypeCache:                         newReqCache,
	TypePortConnector:                 newReqPortConnector,
	TypeSystemSlots:                   newReqSystemSlots,
	TypeOnBoardDevices:                newReqOnBoardDevices,
	TypeOEMStrings:                    newReqOEMStrings,
	TypeSystemConfigurationOptions:    newReqSystemConfigurationOptions,
	TypeBIOSLanguage:                  newReqBIOSLanguage,
	TypeGroupAssociations:             newReqGroupAssociations,
	TypeSystemEventLog:                newReqSystemEventLog,
	TypePhysicalMemoryArray:           newReqPhysicalMemoryArray,
	TypeMemoryDevice:                  newReqMemoryDevice,
	Type32BitMemoryError:              newReq32BitMemoryError,
	TypeMemoryArrayMappedAddress:      newReqMemoryArrayMappedAddress,
	TypeMemoryDeviceMappedAddress:     newReqMemoryDeviceMappedAddress,
	TypeBuiltInPointingDevice:         newReqBuiltInPointingDevice,
	TypePortableBattery:               newReqPortableBattery,
	TypeSystemReset:                   newReqSystemReset,
	TypeHardwareSecurity:              newReqHardwareSecurity,
	TypeSystemPowerControls:           newReqSystemPowerControls,
	TypeVoltageProbe:                  newReqVoltageProbe,
	TypeCoolingDevice:                 newReqCoolingDevice,
	TypeTemperatureProbe:              newReqTemperatureProbe,
	TypeElectricalCurrentProbe:        newReqElectricalCurrentProbe,
	TypeOOBRemoteAccess:               newReqOOBRemoteAccess,
	TypeBootIntegrityServices:         newReqBootIntegrityServices,
	TypeSystemBoot:                    newReqSystemBoot,
	Type64BitMemoryError:              newReq64BitMemoryError,
	TypeManagementDevice:              newReqManagementDevice,
	TypeManagementDeviceComponent:     newReqManagementDeviceComponent,
	TypeManagementDeviceThresholdData: newReqManagementDeviceThresholdData,
	TypeMemoryChannel:                 newReqMemoryChannel,
	TypeIPMIDevice:                    newReqIPMIDevice,
	TypePowerSupply:                   newReqPowerSupply,
	TypeAdditionalInformation:         newReqAdditionalInformation,
	TypeOnBoardDevice:                 newReqOnBoardDevice,
	TypeEndOfTable:                    newReqEndOfTable,
	TypeVendorRangeBegin:              newReqOemSpecificType,
}

func init() {
	handleRegex = regexp.MustCompile(`^Handle\s+(0x[[:xdigit:]]{4}),\s+DMI\s+type\s+(\d+),\s+(\d+)\s+bytes$`)
	inBlockRegex = regexp.MustCompile(`^\t\t(.+)$`)
	recordRegex = regexp.MustCompile(`\t(.+):\s+(.+)$`)
	recordRegex2 = regexp.MustCompile(`\t(.+):$`)
	sizeTypeRegex = regexp.MustCompile(`^(\d+)\s(B|MB|GB)$`)
}

type parsedRecord struct {
	Type ReqType
	Keys map[string]interface{}
}

type records []*parsedRecord

type structTags struct {
	skip      bool
	name      string
	omitempty bool
	omitzero  bool
	comment   string
	commented bool
}

func parseStructTags(tag reflect.StructTag) structTags {
	t := tag.Get("dmidecode")
	if t == "-" {
		return structTags{skip: true}
	}

	var opts structTags
	parts := strings.Split(t, ",")
	opts.name = parts[0]
	for _, s := range parts[1:] {
		switch s {
		case "omitempty":
			opts.omitempty = true
		case "omitzero":
			opts.omitzero = true
		}
	}

	opts.commented, _ = strconv.ParseBool(tag.Get("commented"))
	opts.comment = tag.Get("comment")

	return opts
}

type DMI struct {
	records records
	decoded map[ReqType][]interface{}
}

func New() *DMI {
	return &DMI{
		decoded: make(map[ReqType][]interface{}),
	}
}

func Unmarshal(r io.Reader) (*DMI, error) {
	dmi := New()

	err := dmi.Unmarshal(r)
	if err != nil {
		return nil, err
	}

	return dmi, nil
}

// Unmarshal content in reader into internal tree
func (dmi *DMI) Unmarshal(r io.Reader) error {
	buf, err := ioutil.ReadAll(r)
	if err != nil {
		return fmt.Errorf("dmidecode: read from data stream: %s", err.Error())
	}

	if err = dmi.parse(buf); err != nil {
		return err
	}

	return dmi.decode()
}

// Get by DMI key.
// Unmarshal must be called prior to calling this function
// dst must be in form of *[]S where S is struct Req* from types.go otherwise ErrPtrToSlice is returned
// When function returns nil then dst contains slice with at least one valid entry
// If key is not present in parse data ErrNotFound is returned
func (dmi *DMI) Get(dst interface{}) error {
	dv := reflect.ValueOf(dst)
	if dv.Kind() != reflect.Ptr || dv.IsNil() {
		return ErrPtrToSlice
	}

	dv = dv.Elem()
	reqType, err := checkArg(dv)
	if err != nil {
		return err
	}

	rec, present := dmi.decoded[reqType]
	if !present {
		return ErrNotFound
	}

	// Initialize a slice with Count capacity
	dv.Set(reflect.MakeSlice(dv.Type(), len(rec), len(rec)))

	for i, v := range rec {
		vl := dv.Index(i)
		vl.Set(reflect.Indirect(reflect.ValueOf(v)))
	}

	return nil
}

// Raw return full tree of keys as map[ReqType]interface
func (dmi *DMI) Raw() (interface{}, error) {
	return reflect.Indirect(reflect.ValueOf(dmi.decoded)).Interface(), nil
}

func (dmi *DMI) parse(data []byte) error {
	output := strings.Split(string(data), "\n\n")

	for _, record := range output {
		recordElements := strings.Split(record, "\n")

		// Entries with less than 3 lines are incomplete/inactive; skip them
		if len(recordElements) < 3 {
			continue
		}

		handleData := handleRegex.FindStringSubmatch(recordElements[0])
		if len(handleData) < 3 {
			continue
		}

		var handleID uint16
		if _, err := fmt.Sscan(handleData[1], &handleID); err != nil {
			return fmt.Errorf("dmidecode: parse handle id: %s", err.Error())
		}

		dmiType, err := strconv.Atoi(handleData[2])
		if err != nil {
			return fmt.Errorf("dmidecode: parse record type: %s", err.Error())
		}

		// Loop over the rest of the record, gathering values
		dataElements := recordElements[2:]
		keys := make(map[string]interface{})

		for i := 0; i < len(dataElements); i++ {
			recordData := recordRegex.FindStringSubmatch(dataElements[i])
			// Is this the line containing handle identifier, type, size?
			if len(recordData) > 0 {
				keys[recordData[1]] = recordData[2]
				continue
			}

			// Didn't match regular entry, maybe an array of data?
			recordData = recordRegex2.FindStringSubmatch(dataElements[i])

			if len(recordData) > 0 {
				var arrayValue []string
				keyName := recordData[1]

				for range dataElements[i+1:] {
					inBlockData := inBlockRegex.FindStringSubmatch(dataElements[i+1])
					if len(inBlockData) > 0 {
						arrayValue = append(arrayValue, inBlockData[1])
						i++
					} else {
						break
					}
				}

				keys[keyName] = arrayValue
			}
		}

		dmi.records = append(dmi.records, &parsedRecord{
			Type: ReqType(dmiType),
			Keys: keys,
		})
	}

	return nil
}

func (dmi *DMI) decode() (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("dmidecode: panic %s", r)
		}
	}()

	for _, rec := range dmi.records {
		var obj interface{}

		if rec.Type > 128 {
			rec.Type = 128
		}

		if fn, ok := decoders[rec.Type]; ok {
			obj = fn()
		} else {
			return fmt.Errorf("dmidecode: unknown DMI type: %d", rec.Type)
		}

		rv := indirect(reflect.ValueOf(obj))
		rt := rv.Type()

		for i := 0; i < rt.NumField(); i++ {
			fieldType := rt.Field(i)
			fieldValue := rv.Field(i)

			tag := parseStructTags(fieldType.Tag)
			val := rec.Keys[tag.name]

			if val == nil {
				continue
			}

			if v, ok := val.(string); ok && ((v == "Not Provided") || (v == "Not Specified") || (v == "Unknown")) {
				continue
			}

			switch fieldType.Type.Kind() {
			case reflect.Int:
				vl, e := strconv.Atoi(val.(string))
				if e != nil {
					return fmt.Errorf("dmidecode: parse string into int: %s", e.Error())
				}
				fieldValue.Set(reflect.ValueOf(vl))
			case reflect.Uint16:
				str := val.(string)
				var vl interface{}
				if strings.HasPrefix(str, "0x") {
					v, e := strconv.ParseUint(strings.TrimPrefix(str, "0x"), 16, 16)
					if e != nil {
						return fmt.Errorf("dmidecode: parse hexdecimal: %s", e.Error())
					}

					vl = uint16(v)
				} else {
					v, e := strconv.ParseUint(str, 10, 16)
					if e != nil {
						return fmt.Errorf("dmidecode: parse decimal: %s", e.Error())
					}

					vl = uint16(v)
				}

				fieldValue.Set(reflect.ValueOf(vl))
			case reflect.String:
				// this it annoying bug of dmidecode
				// some of string fields may come as empty thus parser treats them as array
				switch vl := val.(type) {
				case []string:
					if len(vl) > 0 {
						return fmt.Errorf("dmidecode: field expects string but value is slice")
					}

					fieldValue.SetString("")
				case string:
					fieldValue.SetString(vl)
				}
			case reflect.Slice:
				switch fieldValue.Interface().(type) {
				case []byte:
					data, _, e := unifySlice(val, fieldValue)
					if e != nil {
						return fmt.Errorf("dmidecode: unify byte slice: %s", e.Error())
					}

					var bytes []byte
					for i := 0; i < data.Len(); i++ {
						v := data.Index(i).Interface()
						tokens := strings.Split(v.(string), " ")
						for _, token := range tokens {
							decodedByte, e := strconv.ParseUint(token, 16, 8)
							if e != nil {
								return fmt.Errorf("dmidecode: parse hex element into byte: %s", e.Error())
							}

							bytes = append(bytes, byte(decodedByte))
						}
					}

					fieldValue.SetBytes(bytes)
				case []string:
					data, rv1, e := unifySlice(val, fieldValue)
					if e != nil {
						if empty, ok := val.(string); !ok || ((empty != "None") && (empty != "Not Specified")) {
							return fmt.Errorf("dmidecode: unify string slice: %s", err.Error())
						}
					} else {
						for i := 0; i < data.Len(); i++ {
							v := data.Index(i).Interface()
							sliceVal := indirect(rv1.Index(i))
							sliceVal.SetString(v.(string))
						}
					}
				}
			default:
				switch fieldValue.Interface().(type) {
				case time.Time:
					tm, e := time.Parse("01/02/2006", val.(string))
					if e != nil {
						return fmt.Errorf("dmidecode: decode time: %s", e.Error())
					}

					fieldValue.Set(reflect.ValueOf(tm))
				case CPUSignature:
					var sig CPUSignature
					if e := sig.unmarshal(val.(string)); e != nil {
						return fmt.Errorf("dmidecode: decode cpu signature: %s", e.Error())
					}
					fieldValue.Set(reflect.ValueOf(sig))
				case SizeType:
					var vl SizeType

					if e := vl.unmarshal(val.(string)); e != nil {
						return fmt.Errorf("dmidecode: decode size type: %s", e.Error())
					}

					fieldValue.Set(reflect.ValueOf(vl))
				default:
					return fmt.Errorf("dmidecode: unknown type %s", fieldValue.Type().String())
				}
			}
		}

		dmi.decoded[rec.Type] = append(dmi.decoded[rec.Type], obj)
	}
	return nil
}

// indirect returns the value pointed to by a pointer.
// Pointers are followed until the value is not a pointer.
// New values are allocated for each nil pointer.
func indirect(v reflect.Value) reflect.Value {
	if v.Kind() != reflect.Ptr {
		return v
	}
	if v.IsNil() {
		v.Set(reflect.New(v.Type().Elem()))
	}
	return indirect(reflect.Indirect(v))
}

func unifySlice(data interface{}, rv reflect.Value) (reflect.Value, reflect.Value, error) {
	datav := reflect.ValueOf(data)
	if datav.Kind() != reflect.Slice {
		if !datav.IsValid() {
			return reflect.Value{}, reflect.Value{}, nil
		}
		return reflect.Value{}, reflect.Value{}, fmt.Errorf("error")
	}
	n := datav.Len()
	if rv.IsNil() || rv.Cap() < n {
		rv.Set(reflect.MakeSlice(rv.Type(), n, n))
	}
	rv.SetLen(n)

	return datav, rv, nil
}

// checkArg checks that v is type type of []S
//
// It returns what category the slice's elements are, and the reflect.Type
// that represents S.
// spied: https://github.com/StackExchange/wmi
func checkArg(v reflect.Value) (reqType ReqType, err error) {
	var elemType reflect.Type

	if v.Kind() != reflect.Slice {
		return reqType, ErrPtrToSlice
	}

	elemType = v.Type().Elem()
	switch elemType.Kind() {
	case reflect.Struct:
		oType := reflect.New(elemType)
		obj, ok := oType.Interface().(objType)

		if !ok {
			return reqType, ErrInvalidEntityType
		}

		return obj.ObjectType(), nil
	}

	return reqType, ErrPtrToSlice
}
