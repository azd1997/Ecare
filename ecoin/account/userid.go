package account

import (
	"bytes"

	"github.com/azd1997/Ecare/ecoin/utils"
	"github.com/mr-tron/base58"
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
func (userId *UserId) IsValid() (bool, error) {
	fullPubKeyHash, err := base58.Decode(userId.Id)
	if err != nil {
		return false, utils.WrapError("UserID_IsValid", err)
	}
	length := uint(len(fullPubKeyHash))
	actualChecksum := fullPubKeyHash[length-CHECKSUM_LENGTH:]
	version := fullPubKeyHash[0]
	pubKeyHash := fullPubKeyHash[1 : length-CHECKSUM_LENGTH]
	targetChecksum := checksum(append([]byte{version}, pubKeyHash...))
	return bytes.Compare(actualChecksum, targetChecksum) == 0, nil
}
