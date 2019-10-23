# cryptogen

`cryptogen` Hyperledger Fabric の鍵ファイル群を生成するためのユーティリティです。
テスト目的でネットワークを事前設定する手段として提供されます。
通常、実稼働ネットワークの運用では使用されません。## Syntax

``cryptogen`` コマンドには、次の5つのサブコマンドがあります。

  * help
  * generate
  * showtemplate
  * extend
  * version


## cryptogen help
```
usage: cryptogen [<flags>] <command> [<args> ...]

Utility for generating Hyperledger Fabric key material

Flags:
  --help  Show context-sensitive help (also try --help-long and --help-man).

Commands:
  help [<command>...]
    Show help.

  generate [<flags>]
    Generate key material

  showtemplate
    Show the default configuration template

  version
    Show version information

  extend [<flags>]
    Extend existing network


```


## cryptogen generate
```
usage: cryptogen generate [<flags>]

Generate key material

Flags:
  --help                    Show context-sensitive help (also try --help-long
                            and --help-man).
  --output="crypto-config"  The output directory in which to place artifacts
  --config=CONFIG           The configuration template to use

```


## cryptogen showtemplate
```
usage: cryptogen showtemplate

デフォルトの構成テンプレートを表示する

Flags:
  --help  Show context-sensitive help (also try --help-long and --help-man).

```


## cryptogen extend
```
usage: cryptogen extend [<flags>]

既存のネットワークを拡張する

Flags:
  --help                   Show context-sensitive help (also try --help-long and
                           --help-man).
  --input="crypto-config"  The input directory in which existing network place
  --config=CONFIG          The configuration template to use

```


## cryptogen version
```
usage: cryptogen version

バージョン情報を表示する

Flags:
  --help  Show context-sensitive help (also try --help-long and --help-man).

```

## Usage

``cryptogen extend`` コマンドで使用可能なさまざまなフラグを使用した例を次に示します。

```
    cryptogen extend --input="crypto-config" --config=config.yaml

    org3.example.com
```

config.yamlは、 ``org3.example.com`` という新しいピア組織を追加します。

<a rel="license" href="http://creativecommons.org/licenses/by/4.0/"><img alt="Creative Commons License" style="border-width:0" src="https://i.creativecommons.org/l/by/4.0/88x31.png" /></a><br />This work is licensed under a <a rel="license" href="http://creativecommons.org/licenses/by/4.0/">Creative Commons Attribution 4.0 International License</a>.
