package service

const (
	Message_Type_Service_START     	 uint16 = 1
	Message_Type_Service_STARTED     uint16 = 2

	Message_Type_SHUTDOWN  uint16 = 3
	Message_Type_HANDSHAKE uint16 = 4

	Message_Type_POST      uint16 = 10
	Message_Type_GET       uint16 = 11
	Message_Type_PUT       uint16 = 12
	Message_Type_DELETE    uint16 = 13
	Message_Type_PATCH     uint16 = 14
	Message_Type_Request   uint16 = 20
	Message_Type_Reply     uint16 = 21

	Mgmt_Type_Get_Info 	   uint16 = 201
)
