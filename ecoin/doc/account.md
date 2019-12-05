# account包设计

## 前情

在之前的设计中Ecoin的账户体系涉及了以下一些结构体：

- Account: 包含以椭圆曲线算法生成的公钥及私钥，以及后需添加的角色编号属性
- UserId： 包含Account生成的一串字符串作为对外ID，并包含角色编号属性
- Accounts: 个人所持有的Account的集合，负责Account持久化、本地读取等等
- EcoinAccount: 用于记录个人账户的公开属性：余额、挖矿奖励参数、自增参数... 
- EcoinAccounts: 公开账户信息集合，记录各个账户的公开信息

## 原因

为什么设计三层账户体系？

- 与Ecoin的实体角色设计有关。Ecoin定义了A类与B类两大类节点，A类拥有共识权限，需
要维护所有账户的公开信息、状态。

## 重设计

原先的设计高耦合，不利于程序修改。使用接口设计以实现模块解耦，将account模块独立出来

现在需要考虑哪些结构体需要进行抽象化？

1. Account? 我希望以后能够替换Account生成算法，这样导致公私钥长度等信息可能变更，生成的
UserId的Id字符串也可能会改变长度，所以Account需要抽象化为IAccount接口并提供默认实现Account
2. UserId? 为了保证UserId的普适性，规定UserId必须为struct{Id, RoleNo}, Id为
[]byte的类型声明（注意不是类型别名，这样需要显式转换类型）。
3. Accounts? 由于Accounts是个人账户的集合的存储，所以没必要设计复杂结构，所以固定为
struct{[]Account}
4. EcoinAccount？ 存储信息，结构体本身可能有字段变化，但其提供的方法都是Get/Set方法，
没必要设计接口
5. EcoinAccounts? 由于是信息集合，需要被A类节点维护，而且当账户数量增多，这个集合的
存储、查询、新增账户（插入节点）等操作需要尽量保证高效率。初始先直接使用map，后期需要
选择、设计更为合适的数据结构。因此这个结构需要进行抽象化
6. 从前5点来看，原先的账户模型拆分为account、ecoinaccount两个包比较合适，
ecoinaccount包负责EcoinAccount相关。

## 具体属性方法设计

account包主要实现Account。

虽然Account关联了许多方法，比如说创建新交易等等，但是出于尽量将包与包之间的耦合降到最低的想法，
Account包未依赖其他任何核心包

只引入了utils和log，未来还会引入logger。但其他不可能再有了。

另外，全局配置将用在真正无法解耦的中央模块，接近应用入口的层面，这里不会引入。

### 