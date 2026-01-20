
```text
{vFloat . "companyId" "companyId.not.Int" "公司ID必须是数子" "params.companyId"}
{vRequired . "companyName" "companyName.required" "公司名必须" "params.companyName"}
{vReg . "^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\\.[A-Za-z]{2,}$" "email" "email.required" "邮箱格式不对" "$.params.email"}
{vRequired . "email2" "email2.required" "email2名必须" "params.email2"}

select count(*) from common_address where company_id = {val . "params.companyId"};
```

```text
{val . "path1", "path2", 'pathN'}
{expr . "a.goods" "=" "company.goods"}
{getValue "path1", "path2", 'pathN'}
```

```go
import (
_ "github.com/qq1060656096/drugo-provider/biapi/api"
)

```