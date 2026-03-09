# API Purpose

Since this repo is about a Hyperledger Fabric network and a TEE Auction Container, this go api in this folder is designed to simulate the government's trusted databases.

# Mock Databases

Mock databases are used to simulate the government's trusted databases, which provide essential data for the carbon credit market. These databases include:

## Sicar API

From https://docs.dataprev.gov.br/docs/api/

### CPF/CNPJ

https://papisp.dataprev.gov.br/sicar/cpfcnpj/1.0/{cpfCnpj}
GET /{cpfCnpj}

cpfCnpj: CPF ou CNPJ do proprietário do imóvel registrado no CAR
string (path)

```json
Response:
{
  "result": [
    {
      "identificadorimovel": 99999999,
      "codigoimovel": "UF-9999999-01234567890ABCDEFGHIJKLMNOPQRSTU",
      "ativo": true
    }
  ]
}
```

### Demonstrativo

https://papisp.dataprev.gov.br/sicar/demonstrativoDegustacao/1.0/{codigoImovel}
GET /{codigoImovel}

codigoImovel: Código do imóvel registrado no CAR
string (path)

Response:

```json
{
  "result": [
    {
        "situacaoImovel":"PE",
        "codigoImovel":"SC-4210100-EA1F3824FC3D4148B83F921E253C6DCD",
        "descricaoEtapaCadastro":"Aguardando análise, não passível de revisão de dados",
        "areaTotalImovel":95.5569,
        "quantidadeModulosFiscais":"5.9723",
        "dataCadastro":"12/06/2015",
        "dataUltimaAtualizacaoCadastro":"01/12/2022",
        "codigoMunicipio":2902005,
        "municipio":"Mafra",
        "unidadeFederativa":"SC",
        "coordenadaImovelX":-49.6610956240856,
        "coordenadaImovelY":-26.3179762554547,
        "areaRemanescenteVegetacaoNativa":32.8631417737116,
        "areaConsolidada":62.634481425266,
        "areaServidaoAdministrativa":0,
        "situacaoReservaLegal":"Não analisada",
        "areaReservaLegalAverbada":"0.0000",
        "areaReservaLegalAprovadaNaoAverbada":"0.0000",
        "areaReservaLegalProposta":"11.9533",
        "areaReservaLegalDeclaradaProprietarioPossuidor":12.8,
        "areaPreservacaoPermanente":16.5969602125317,
        "areaPreservacaoPermanenteAreaRuralConsolida":"0.1422",
        "areaPreservacaoPermanenteAreaRemanescenteVegetacaoNativa":"16.4547",
        "areaUsoRestritoDeclividade":0,
        "areaReservaLegalExcedentePassivo":7.15,
        "areaUsoRestrito":0,
        "areaReservaLegalRecompor":0,
        "areaPreservacaoPermanenteRecompor":0.07,
        "areaUsoRestritoRecompor":0,
        "sobreposicoesTerraIndigena":0,
        "sobreposicoesUnidadeConservacao":0,
        "sobreposicoesAreasEmbargadas":0.4086,
        "poligonoAreaImovel": "MULTIPOLYGON(((-99.99999999 -00.00000000,-99.99999999 -99.99999999,-00.00000000 -99.99999999,-00.00000000 -00.00000000,-99.99999999 -00.00000000)))"
    }
  ]
}
```

### Sicar Imóvel

https://papisp.dataprev.gov.br/sicar/pra/1.0/{codigoImovel}
GET /{codigoImovel}

codigoImovel: Código do imóvel registrado no CAR
string (path)

Response:

```json
{
  "result": [
    {
      "identificadorimovel": 2345678,
      "codigoimovel": "UF-9999999-A38E4A53E2494858B5FD8386DF178392",
      "codigoversao": "1.9.2",
      "tipoorigem": "EST",
      "protocolo": "UF-1234567-A38E4A53E2494858B5FD8386DF178392",
      "dataultimoprotocolo": "01/01/2024",
      "statusimovel": "RE",
      "tipoimovel": "IRU",
      "cpfcadastrante": "01234567890",
      "nomecadastrante": "JOSE CADASTRANTE",
      "nomeimovel": "FAZENDA DO CADASTRANTE",
      "fracaoideal": null,
      "identificadormunicipio": 5102504,
      "municipio": "Cáceres",
      "areatotalimovel": "1234.5678",
      "numeromodulosfiscais": "12.3456",
      "datacriacaoregistro": "01/01/2024",
      "dataatualizacaoregistro": "01/06/2024",
      "areatotal": "MULTIPOLYGON(((-99.99999999 -00.00000000,-99.99999999 -99.99999999,-00.00000000 -99.99999999,-00.00000000 -00.00000000,-99.99999999 -00.00000000)))",
      "declaracaoanterior": null,
      "declaracaoposterior": 1234567,
      "situacao": false,
      "situacaomigracao": true,
      "aderiupra": "Não"
    }
  ]
}
```

### Tema

https://papisp.dataprev.gov.br/sicar/tema/1.0/{codigoImovel}
GET /{codigoImovel}

codigoImovel: Código do imóvel registrado no CAR
string (path)

Response:
```json
{
  "tema": "Área Líquida do Imóvel",
  "identificadorImovel": "2345678",
  "areaTotalTema": "1.23456789012345",
  "geoAreaTema": "MULTIPOLYGON(((-99.99999999 -00.00000000,-99.99999999 -99.99999999,-00.00000000 -99.99999999,-00.00000000 -00.00000000,-99.99999999 -00.00000000)))"
}
```

### Atualização de Imóvel

https://papisp.dataprev.gov.br/sicar/generico-imovel-atualizacao/1.0/{data}

GET /{data}

data: Data da atualização no formato DDMMAAAA
string (path)

Response:

```json
{
  "result": [
    {
      "identificadorImovel": 665124,
      "codigoImovel": "PR-4100301-CE8A8541F33C4BD788F95403E69827DC",
      "dataCriacaoCadastro": "02/05/2025",
      "dataUltimaAtualizacaoCadastro": "02/05/2025"
    }
  ]
}```


### Recibo

https://papisp.dataprev.gov.br/sicar/recibo/1.0/{codigoImovel}

GET/{codigoImovel}

codigoImovel: Código do imóvel registrado no CAR
string (path)

Response:

```json
{
  "result": [
    {
      "identificadorImovel": 12579165,
      "codigoImovel": "TO-1714203-1C6DC93596CE4193A5E8519DDED552A4",
      "situacaoImovel": "RE",
      "tipoImovel": "AST",
      "dataCadastro": "",
      "nomeImovel": "JACUBINHA",
      "codigoMunicipio": 1714203,
      "municipio": "Natividade",
      "unidadefederativa": "TO",
      "coordenadaImovelX": -51.8516375657968,
      "coordenadaImovelY": 0.282749156200535,
      "areaTotalImovel": "6326.184",
      "moduloFiscal": "1.1297",
      "protocolo": "TO-1714203-037F2BE93ECB15CA6F2D256AAD97F2FB",
      "informacoesAdicionais": "Foi detectada uma diferença entre a área do imóvel rural declarada conforme documentação comprobatória de propriedade/posse/concessão [1111.0 hectares] e a área do imóvel rural identificada em representação gráfica [1.111,7294 hectares].",
      "geoImovel": "MULTIPOLYGON(((-51.865440320485 0.298197106149128,-51.8377743427669 0.299603850778863,-51.8373054278903 0.267717639171532,-51.8659092353616 0.265841979665218,-51.865440320485 0.298197106149128)))",
      "proprietarios": {
        "tipoPessoa": "PF",
        "cpfCnpj": "03662243601",
        "nomeProprietario": "Pessoa teste SFB",
        "nomeFantasia": "Empresa teste SFB"
      },
      "areaServidaoAdministrativa": 0,
      "areaLiquidaImovel": 0,
      "areaPreservacaoPermanente": 909.1799,
      "areaUsoRestrito": 0,
      "areaConsolidada": 1675.4077,
      "areaRemanescenteVegetacaoNativa": 4067.6421,
      "areaReservaLegal": 2182.3045,
      "matricula": "M-2.096",
      "dataMatricula": "",
      "livroMatricula": "2-J",
      "folhaMatricula": "79",
      "municipioCartorio": "Natividade",
      "ufCartorio": "TO"
    }
  ]
}
```

### Endereço

https://papisp.dataprev.gov.br/sicar/generico-imovel-endereco/1.0/{codigoImovel}

GET/{codigoImovel}

codigoImovel: Código do imóvel registrado no CAR
string (path)

Response:
```json
{
  "result": [
    {
      "idt_endereco_imovel": 635491,
      "idt_imovel": 664856,
      "idt_municipio": 2601706,
      "nom_logradouro": "av",
      "num_endereco": "011",
      "des_complemento": null,
      "nom_bairro": "centro",
      "cod_cep": "23689334",
      "des_telefone": null,
      "des_email": null
    }
  ]
}
```

## Generico Imovel Pessoa

https://papisp.dataprev.gov.br/sicar/generico-imovel-pessoa/1.0/{codigoImovel}
GET/{codigoImovel}

codigoImovel: Código do imóvel registrado no CAR
string (path)

Response:

```json
{
  "result": [
    {
      "idt_imovel_pessoa": 990021,
      "idt_imovel": 664856,
      "ind_tipo_pessoa": "PF",
      "cod_cpf_cnpj": "08903303440",
      "nom_completo": "Teste MTC",
      "dat_nascimento": "1992-11-23",
      "nom_mae": "Mãe teste",
      "cod_cpf_conjuge": null,
      "nom_completo_conjuge": null,
      "nom_fantasia": null
    }
  ]
}
```

### Generico Imovel Tema Poligono

https://papisp.dataprev.gov.br/sicar/generico-imovel-tema-poligono/1.0/{codigoImovel}
GET/{codigoImovel}

codigoImovel: Código do imóvel registrado no CAR
string (path)

Response:

```json
{
  "result": [
    {
      "idt_rel_tema_imovel": 9257616,
      "idt_tema": 2,
      "idt_imovel": 664856,
      "num_area": 1.29985632917583,
      "wkt_shape": "MULTIPOLYGON(((-36.3974864796876 -8.30181329909558,-36.3979303836822 -8.30188879561361,-36.3983995936993 -8.29971683165591,-36.3983995936993 -8.29971683165591,-36.3978016376495 -8.29985043981347,-36.3974864796876 -8.30181329909558)))"
    }
  ]
}
```

### Generico Imovel Tema Ponto

https://papisp.dataprev.gov.br/sicar/generico-imovel-tema-ponto/1.0/{codigoImovel}

GET/{codigoImovel}

codigoImovel: Código do imóvel registrado no CAR
string (path)

Response:

```json
{
  "result": [
    {
      "idt_rel_tema_imovel": 9257629,
      "idt_tema": 15,
      "idt_imovel": 664856,
      "num_coordenada_x": -36.3970291614532,
      "num_coordenada_y": -8.30114564628698,
      "wkt_shape": "POINT(-36.3970291614532 -8.30114564628698)"
    }
  ]
}
```

### Generico Imovel

https://papisp.dataprev.gov.br/sicar/generico-imovel/1.0/{codigoImovel}
GET/{codigoImovel}

codigoImovel: Código do imóvel registrado no CAR
string (path)

Response:

```json
{
  "result": [
    {
      "idt_imovel": 664856,
      "cod_imovel": "PE-2601706-99F8DAC4CF714DCB82E62BEF4079EA46",
      "cod_versao": "3.4",
      "ind_tipo_origem": "OFF",
      "cod_protocolo": "PE-2601706-84B8FA1903BFC7590B88B36E57ED497D",
      "dat_protocolo": "27/01/2025",
      "ind_status_imovel": "PE",
      "ind_tipo_imovel": "IRU",
      "cod_cpf_cadastrante": "08903303440",
      "nom_completo_cadastrante": "Teste MTC",
      "nom_imovel": "Teste_MTC_MRA",
      "num_fracao_ideal": null,
      "idt_municipio": 2601706,
      "num_area_imovel": "11.3746",
      "num_modulo_fiscal": "0.5687",
      "dat_criacao": "27/01/2025",
      "dat_atualizacao": "01/02/2025",
      "st_astext": "MULTIPOLYGON(((-36.3965570926666 -8.30165523454822,-36.3975441455841 -8.29564629756826,-36.3991963863373 -8.2960284941768,-36.3979303836822 -8.30188879561361,-36.3965570926666 -8.30165523454822)))",
      "idt_imovel_anterior": null,
      "idt_imovel_posterior": null,
      "flg_ativo": true,
      "flg_migracao": false
    }
  ]
}
```

### Mir

https://papisp.dataprev.gov.br/sicar/mir/1.0/{cpfCnpj}

GET/{cpfCnpj}

cpfCnpj: CPF ou CNPJ do proprietário do imóvel registrado no CAR
string (path)

Response:

```json
{
  "result": [
    {
      "identificadorImovel": "3936065",
      "codigoImovel": "BA-2902104-3FE0E82BF94040EDA65C92AD48E19107",
      "indicadorAtivo": true
    }
  ]
}
```

# Mock data

From the responses specs, we must have a function that creates mock data for the APIs. All apis well be emulated in this single one.

# Certficate

We need to generate a certificate if the file path is not available. This certificate will be used for blockchain and TEE to validate the authenticity of the data.
