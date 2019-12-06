package account

import (
	"bytes"
	"github.com/mr-tron/base58"

	"github.com/azd1997/Ecare/ecoin/utils"
)

// UserId 用户身份标识符，包含标识符和角色编号两个属性
type UserId struct {
	Id     string `json:"id"`
	RoleNo uint   `json:"roleNo"` // 角色编号，参见role.go
}

// NewUserId 根据Id生成UserId
func NewUserId(id string, roleNo uint) *UserId {
	// 检查Id是否已存在，由上层进行
	return &UserId{
		Id:     id,
		RoleNo: roleNo,
	}
}

// String 转换为json字符串
func (userId *UserId) String() string {
	return utils.JsonMarshalIndentToString(userId)
}

// IsValid 判断UserId.Id是否有效
func (userId *UserId) IsValid() error {
	fullPubKeyHash, err := base58.Decode(userId.Id)
	if err != nil {
		return utils.WrapError("UserID_IsValid", err)
	}
	length := uint(len(fullPubKeyHash))
	actualChecksum := fullPubKeyHash[length-CHECKSUM_LENGTH:]
	version := fullPubKeyHash[0]
	pubKeyHash := fullPubKeyHash[1 : length-CHECKSUM_LENGTH]
	targetChecksum := checksum(append([]byte{version}, pubKeyHash...))
	if bytes.Compare(actualChecksum, targetChecksum) != 0 {		// 不相等
		return utils.WrapError("UserID_IsValid", ErrInvalidUserId)
	}

	return nil
}

// RoleOk 判断角色是否符合条件。默认情况下0~9为A类角色，10~99为B类角色
// 当pattern设定为大类型查询之后，role不被使用，随便设
func (userId *UserId) RoleOk(pattern uint8, role uint) error {
	switch pattern {
	case A:
		if userId.RoleNo >= 0 && userId.RoleNo <= 9 {return nil}
		return ErrNotRoleA
	case B:
		if userId.RoleNo >= 10 && userId.RoleNo < 100 {return nil}
		return ErrNotRoleB
	case All:
		if userId.RoleNo >= 0 && userId.RoleNo < 100 {return nil}
		return ErrUnKnownRole
	case Single:
		if userId.RoleNo == role {return nil}
		return ErrUnmatchedRole
	default:
		return ErrUnKnownQueryPattern
	}
}

