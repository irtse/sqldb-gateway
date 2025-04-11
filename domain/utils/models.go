package utils

type AbstractDomain struct {
	TableName  string
	AutoLoad   bool
	User       string
	UserID     string
	Shallowed  bool
	SuperAdmin bool
	RawView    bool
	Super      bool
	Empty      bool
	LowerRes   bool
	Own        bool
	Method     Method
	Params     Params
}

func (d *AbstractDomain) SetOwn(own bool) {
	d.Own = own
}
func (d *AbstractDomain) GetMethod() Method { return d.Method }
func (d *AbstractDomain) GetEmpty() bool    { return d.Empty }
func (d *AbstractDomain) GetUserID() string {
	return d.UserID
}
func (d *AbstractDomain) GetUser() string     { return d.User }
func (d *AbstractDomain) IsSuperAdmin() bool  { return d.SuperAdmin }
func (d *AbstractDomain) IsSuperCall() bool   { return d.Super && d.SuperAdmin }
func (d *AbstractDomain) IsShallowed() bool   { return d.Shallowed }
func (d *AbstractDomain) GetParams() Params   { return d.Params }
func (d *AbstractDomain) GetTable() string    { return d.TableName }
func (d *AbstractDomain) IsLowerResult() bool { return d.LowerRes }
func (d *AbstractDomain) IsOwn(checkPerm bool, force bool, method Method) bool {
	return d.Own
}
