package uuid

import uuid "github.com/satori/go.uuid"

const uuidSize = 16

type UUID [uuidSize]byte

var Nil = UUID{}

func (u *UUID) Scan(src interface{}) error {
	var impl uuid.UUID
	err := impl.Scan(src)

	*u = UUID(impl)
	return err
}

func (u UUID) String() string {
	impl := uuid.UUID(u)
	return impl.String()
}

func (u UUID) Bytes() []byte {
	impl := uuid.UUID(u)
	return impl.Bytes()
}

func (u *UUID) UnmarshalText(text []byte) error {
	var impl uuid.UUID
	err := impl.UnmarshalText(text)

	*u = UUID(impl)
	return err
}

func (u UUID) MarshalText() ([]byte, error) {
	impl := uuid.UUID(u)
	return impl.MarshalText()
}

func FromString(input string) (u UUID, err error) {
	impl, err := uuid.FromString(input)
	if err != nil {
		return u, err
	}
	u = UUID(impl)
	return
}

func ToStrings(uuids []UUID) []string {
	result := make([]string, 0, len(uuids))
	for _, uid := range uuids {
		result = append(result, uid.String())
	}
	return result
}

func FromStrings(ids []string) ([]UUID, error) {
	result := make([]UUID, 0, len(ids))
	for _, id := range ids {
		uid, err := FromString(id)
		if err != nil {
			return []UUID{}, err
		}
		result = append(result, uid)
	}
	return result, nil
}
