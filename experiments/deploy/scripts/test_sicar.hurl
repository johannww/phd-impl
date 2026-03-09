GET {{base_url}}/sicar/pra/1.0/{{codigo_imovel}}
HTTP 200
[Asserts]
jsonpath "$.result[0].codigoimovel" == "{{codigo_imovel}}"
jsonpath "$.result[0].areatotalimovel" isString
jsonpath "$.result[0].municipio" isString
jsonpath "$.result[0].dataultimoprotocolo" isString

GET {{base_url}}/sicar/demonstrativoDegustacao/1.0/{{codigo_imovel}}
HTTP 200
[Asserts]
jsonpath "$.result[0].codigoImovel" == "{{codigo_imovel}}"
jsonpath "$.result[0].areaTotalImovel" isFloat
jsonpath "$.result[0].municipio" isString
jsonpath "$.result[0].dataCadastro" isString

GET {{base_url}}/sicar/recibo/1.0/{{codigo_imovel}}
HTTP 200
[Asserts]
jsonpath "$.result[0].codigoImovel" == "{{codigo_imovel}}"
jsonpath "$.result[0].areaTotalImovel" isString
jsonpath "$.result[0].municipio" isString
jsonpath "$.result[0].dataCadastro" isString
