package speed

import (
	"fmt"
	"sync"
)

// MetricType is an enumerated type representing all valid types for a metric
type MetricType int32

// Possible values for a MetricType
const (
	NoSupportType MetricType = iota
	Int32Type
	Uint32Type
	Int64Type
	Uint64Type
	FloatType
	DoubleType
	StringType
	AggregateType
	AggregateStaticType
	EventType
	HighresEventType
	UnknownType
)

//go:generate stringer -type=MetricType

// MetricUnit is an enumerated type representing all possible values for a valid PCP unit
type MetricUnit int32

// SpaceUnit is an enumerated type representing all units for space
type SpaceUnit MetricUnit

// Possible values for SpaceUnit
const (
	ByteUnit SpaceUnit = iota
	KilobyteUnit
	MegabyteUnit
	GigabyteUnit
	TerabyteUnit
	PetabyteUnit
	ExabyteUnit
)

//go:generate stringer -type=SpaceUnit

// TimeUnit is an enumerated type representing all possible units for representing time
type TimeUnit MetricUnit

// Possible Values for TimeUnit
const (
	NanosecondUnit TimeUnit = iota
	MicrosecondUnit
	MillisecondUnit
	SecondUnit
	MinuteUnit
	HourUnit
)

//go:generate stringer -type=TimeUnit

// CountUnit is a type representing a counted quantity
type CountUnit MetricUnit

// OneUnit represents the only CountUnit
const OneUnit CountUnit = iota

//go:generate stringer -type=CountUnit

// MetricSemantics represents an enumerated type representing the possible
// values for the semantics of a metric
type MetricSemantics int32

// Possible values for MetricSemantics
const (
	NoSemantics MetricSemantics = iota
	CounterSemantics
	InstantSemantics
	DiscreteSemantics
)

//go:generate stringer -type=MetricSemantics

// Metric defines the general interface a type needs to implement to qualify
// as a valid PCP metric
type Metric interface {
	Val() interface{}           // gets the value of the metric
	Set(interface{}) error      // sets the value of the metric to a value, optionally returns an error on failure
	ID() uint32                 // gets the unique id generated for this metric
	Name() string               // gets the name for the metric
	Type() MetricType           // gets the type of a metric
	Unit() MetricUnit           // gets the unit of a metric
	Semantics() MetricSemantics // gets the semantics for a metric
	Description() string        // gets the description of a metric
}

// PCPMetricItemBitLength is the maximum bit size of a PCP Metric id
//
// see: https://github.com/performancecopilot/pcp/blob/master/src/include/pcp/impl.h#L102-L121
const PCPMetricItemBitLength = 10

// MetricDesc is a metric metadata wrapper
// each metric type can wrap its metadata by containing a MetricDesc type and only define its own
// specific properties assuming MetricDesc will handle the rest
//
// when writing, this type is supposed to map directly to the pmDesc struct as defined in PCP core
type MetricDesc struct {
	id                                uint32          // unique metric id
	name                              string          // the name
	indom                             InstanceDomain  // the instance domain
	t                                 MetricType      // the type of a metric
	sem                               MetricSemantics // the semantics
	u                                 MetricUnit      // the unit
	offset                            int             // memory storage offset for the metric description
	shortDescription, longDescription string
}

// NewMetricDesc creates a new Metric Description wrapper type
func NewMetricDesc(n string, i InstanceDomain, t MetricType, s MetricSemantics, u MetricUnit, short, long string) *MetricDesc {
	return &MetricDesc{
		getHash(n, PCPMetricItemBitLength), n, i, t, s, u, 0, short, long,
	}
}

func (md *MetricDesc) Offset() int { return md.offset }

func (md *MetricDesc) SetOffset(offset int) { md.offset = offset }

func (md *MetricDesc) String() string {
	return fmt.Sprintf("%s{%v, %v, %v, %v}", md.name, md.indom, md.t, md.sem, md.u)
}

// PCPMetric defines a PCP compatible metric type that can be constructed by specifying values
// for type, semantics and unit
type PCPMetric struct {
	sync.RWMutex
	val    interface{} // all bets are off, store whatever you want
	desc   *MetricDesc // the metadata associated with this metric
	offset int         // memory storage offset for the metric value
}

// NewPCPMetric creates a new instance of PCPMetric
func NewPCPMetric(val interface{}, name string, indom InstanceDomain, t MetricType, s MetricSemantics, u MetricUnit, short, long string) *PCPMetric {
	return &PCPMetric{
		val:  val,
		desc: NewMetricDesc(name, indom, t, s, u, short, long),
	}
}

// Val returns the current set value of PCPMetric
func (m *PCPMetric) Val() interface{} {
	m.RLock()
	defer m.RUnlock()
	return m.val
}

// Set sets the current value of PCPMetric
func (m *PCPMetric) Set(val interface{}) error {
	if val != m.val {
		m.Lock()
		defer m.Unlock()
		m.val = val
	}
	return nil
}

// ID returns the generated id for PCPMetric
func (m *PCPMetric) ID() uint32 { return m.desc.id }

// Name returns the generated id for PCPMetric
func (m *PCPMetric) Name() string { return m.desc.name }

// Semantics returns the current stored value for PCPMetric
func (m *PCPMetric) Semantics() MetricSemantics { return m.desc.sem }

// Unit returns the unit for PCPMetric
func (m *PCPMetric) Unit() MetricUnit { return m.desc.u }

// Type returns the type for PCPMetric
func (m *PCPMetric) Type() MetricType { return m.desc.t }

// Description returns the description for PCPMetric
func (m *PCPMetric) Description() string {
	sd := m.desc.shortDescription
	ld := m.desc.longDescription
	if len(ld) > 0 {
		return sd + "\n\n" + ld
	}
	return sd
}

func (m *PCPMetric) Offset() int { return m.offset }

func (m *PCPMetric) SetOffset(offset int) { m.offset = offset }

func (m *PCPMetric) String() string {
	return fmt.Sprintf("Val: %v\n%v", m.val, m.Description())
}

// TODO: implement PCPCounterMetric, PCPGaugeMetric ...
