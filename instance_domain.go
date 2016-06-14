package speed

import "errors"

// InstanceDomain wraps a PCP compatible instance domain
type InstanceDomain struct {
	id                          uint32
	name                        string
	instances                   map[uint32]*Instance // the instances for this InstanceDomain stored as a map
	shortHelpText, longHelpText string
}

// NOTE: this declaration alone doesn't make this usable
// it needs to be 'made' at the beginning of monitoring
var instanceDomains map[uint32]*InstanceDomain

// NewInstanceDomain creates a new instance domain or returns an already created one for the passed name
// NOTE: this is different from parfait's idea of generating ids for InstanceDomains
// We simply generate a unique 32 bit hash for an instance domain name, and if it has not
// already been created, we create it, otherwise we return the already created version
func NewInstanceDomain(name, shortDescription, longDescription string) *InstanceDomain {
	h := getHash(name)

	v, present := instanceDomains[h]
	if present {
		return v
	}

	cache := make(map[uint32]*Instance)
	instanceDomains[h] = &InstanceDomain{
		id:            h,
		name:          name,
		instances:     cache,
		shortHelpText: shortDescription,
		longHelpText:  longDescription,
	}

	return instanceDomains[h]
}

// AddInstance adds a new instance to the current InstanceDomain
func (indom *InstanceDomain) AddInstance(name string) error {
	h := getHash(name)

	_, present := indom.instances[h]
	if present {
		return errors.New("Instance with same name already created for the InstanceDomain")
	}

	ins := newInstance(h, name, indom)
	indom.instances[h] = ins

	return nil
}

func (indom *InstanceDomain) String() string {
	s := "InstanceDomain: " + indom.name
	if len(indom.instances) > 0 {
		s += "["
		for _, i := range indom.instances {
			s += i.name + ","
		}
		s += "]"
	}
	return s
}
