# Identity

## What is an Identity?

The different actors in a blockchain network include peers, orderers, client
applications, administrators and more. Each of these actors --- active elements
inside or outside a network able to consume services --- has a digital identity
encapsulated in an X.509 digital certificate. These identities really matter
because they **determine the exact permissions over resources and access to
information that actors have in a blockchain network.**

ブロックチェーンネットワークのさまざまなアクターには、ピア、注文者、クライアントアプリケーション、管理者などが含まれます。
これらの各アクター（サービスを消費できるネットワークの内部または外部のアクティブな要素）には、X.509デジタル証明書にカプセル化されたデジタルIDがあります。
これらのIDは、**リソースに対する正確な許可と、アクターがブロックチェーンネットワーク内で持っている情報へのアクセスを決定するため**、とてもに重要です。


A digital identity furthermore has some additional attributes that Fabric uses
to determine permissions, and it gives the union of an identity and the associated
attributes a special name --- **principal**. Principals are just like userIDs or
groupIDs, but a little more flexible because they can include a wide range of
properties of an actor's identity, such as the actor's organization, organizational
unit, role or even the actor's specific identity. When we talk about principals,
they are the properties which determine their permissions.

さらに、デジタルIDには、Fabricがアクセス許可を決定するために使用するいくつかの追加属性があり、IDと関連属性の結合に特別な名前 **「プリンシパル」** を与えます。
プリンシパルは、ユーザーIDまたはグループIDに似ていますが、アクターの組織、組織単位、役割、またはアクターの特定のアイデンティティなど、アクターのアイデンティティの幅広いプロパティを含めることができるため、もう少し柔軟です。
ここでプリンシパルについて話すときは、パーミッションを決定するプロパティとします。

For an identity to be **verifiable**, it must come from a **trusted** authority.
A [membership service provider](../membership/membership.html)
(MSP) is how this is achieved in Fabric. More specifically, an MSP is a component
that defines the rules that govern the valid identities for this organization.
The default MSP implementation in Fabric uses X.509 certificates as identities,
adopting a traditional Public Key Infrastructure (PKI) hierarchical model (more
on PKI later).

IDを**検証可能にする**には、**信頼できる機関**から取得する必要があります。
[メンバーシップサービスプロバイダー](../membership/membership.html)（MSP:Membership Service Provider）は、Fabricでこれを実現する方法です。
より具体的には、MSPは、この組織の有効なIDを管理するルールを定義するコンポーネントです。
FabricのデフォルトのMSP実装では、X.509証明書をIDとして使用し、従来の公開鍵暗号基盤（PKI）階層モデルを採用しています（PKIについては後ほど説明します）。

## A Simple Scenario to Explain the Use of an Identity

Imagine that you visit a supermarket to buy some groceries. At the checkout you see
a sign that says that only Visa, Mastercard and AMEX cards are accepted. If you try to
pay with a different card --- let's call it an "ImagineCard" --- it doesn't matter whether
the card is authentic and you have sufficient funds in your account. It will be not be
accepted.

スーパーに行って食料品を買うと想像してください。
チェックアウト時に、Visa、Mastercard、AMEXカードのみが受け入れられるというサインが表示されます。
別のカードで支払いをしようとする場合（「ImagineCard」と呼びましょう）、カードが本物であり、アカウントに十分な資金があるかどうかは関係ありません。 
支払いは受け入れられないでしょう。

![Scenario](./identity.diagram.6.png)

*Having a valid credit card is not enough --- it must also be accepted by the store! PKIs
and MSPs work together in the same way --- a PKI provides a list of identities,
and an MSP says which of these are members of a given organization that participates in
the network.*

*有効なクレジットカードを持っているだけでは十分ではありません。
ストアで承認する必要もあります。 
PKIとMSPは同じように連携します。
PKIはIDのリストを提供し、MSPはこれらのどれがネットワークに参加する特定の組織のメンバーであるかを示します。*

PKI certificate authorities and MSPs provide a similar combination of functionalities.
A PKI is like a card provider --- it dispenses many different types of verifiable
identities. An MSP, on the other hand, is like the list of card providers accepted
by the store, determining which identities are the trusted members (actors)
of the store payment network. **MSPs turn verifiable identities into the members
of a blockchain network**.

Let's drill into these concepts in a little more detail.

PKI認証局とMSPは、同様の機能の組み合わせを提供します。
PKIはカードプロバイダーのようなもので、さまざまな種類の検証可能なIDを発行します。
一方、MSPは、店舗が受け入れるカードプロバイダーのリストのようなもので、どのIDが店舗決済ネットワークの信頼できるメンバー（アクター）であるかを判断します。
**MSPは、検証可能なIDをブロックチェーンネットワークのメンバーに変えます。**

これらの概念をもう少し詳しく見ていきましょう。

## What are PKIs?

**A public key infrastructure (PKI) is a collection of internet technologies that provides
secure communications in a network.** It's PKI that puts the **S** in **HTTPS** --- and if
you're reading this documentation on a web browser, you're probably using a PKI to make
sure it comes from a verified source.

**公開鍵暗号基盤（PKI）は、ネットワークで安全な通信を提供するインターネット技術の集まりです。 **
**HTTPS**を**S**の状態にするのはPKIです。
このドキュメントをWebブラウザーで読んでいる場合は、PKIを使用して、検証済みのソースからのものであることを確認している可能性があります。

![PKI](./identity.diagram.7.png)

*The elements of Public Key Infrastructure (PKI). A PKI is comprised of Certificate
Authorities who issue digital certificates to parties (e.g., users of a service, service
provider), who then use them to authenticate themselves in the messages they exchange
with their environment. A CA's Certificate Revocation List (CRL) constitutes a reference
for the certificates that are no longer valid. Revocation of a certificate can happen for
a number of reasons. For example, a certificate may be revoked because the cryptographic
private material associated to the certificate has been exposed.*

*公開キー基盤（PKI）の要素。PKIは、関係者（たとえば、サービスのユーザー、サービスプロバイダー）にデジタル証明書を発行する認証局で構成されます。
認証局は、デジタル証明書を使用して、彼らの環境で交換するメッセージにおいて自身を認証します。
CAの証明書失効リスト（CRL:Certificate Revocation List）は、無効になった証明書の参照先で構成されています。
証明書の失効は、さまざまな理由で発生する可能性があります。たとえば、証明書に関連付けられている暗号化された個人情報が公開されているため、証明書が取り消される場合があります。*

Although a blockchain network is more than a communications network, it relies on the
PKI standard to ensure secure communication between various network participants, and to
ensure that messages posted on the blockchain are properly authenticated.
It's therefore important to understand the basics of PKI and then why MSPs are
so important.

ブロックチェーンネットワークは通信ネットワーク以上のものですが、さまざまなネットワーク参加者間の安全な通信を確保し、ブロックチェーンに投稿されたメッセージが適切に認証されるようにするために、PKI標準に依存しています。
したがって、PKIの基本を理解し、次にMSPが非常に重要である理由を理解することが重要です。

There are four key elements to PKI:

 * **Digital Certificates**
 * **Public and Private Keys**
 * **Certificate Authorities**
 * **Certificate Revocation Lists**

Let's quickly describe these PKI basics, and if you want to know more details,
[Wikipedia](https://en.wikipedia.org/wiki/Public_key_infrastructure) is a good
place to start.

PKIには4つの重要な要素があります。

 * **デジタル証明書**
 * **公開鍵と秘密鍵**
 * **認証局**
 * **証明書失効リスト**
 
これらのPKIの基本を簡単に説明しましょう。
詳細を知りたい場合は、[ウィキペディア](https://ja.wikipedia.org/wiki/%E5%85%AC%E9%96%8B%E9%8D%B5%E5%9F%BA%E7%9B%A4)が始めるのにちょうどいい場所になるでしょう。

## デジタル証明書（Digital Certificates）

A digital certificate is a document which holds a set of attributes relating to
the holder of the certificate. The most common type of certificate is the one
compliant with the [X.509 standard](https://en.wikipedia.org/wiki/X.509), which
allows the encoding of a party's identifying details in its structure.

デジタル証明書は、証明書の所有者に関する一連の属性を保持するドキュメントです。
最も一般的なタイプの証明書は、[X.509標準](https://ja.wikipedia.org/wiki/X.509)に準拠している証明書であり、その構造内で当事者の識別詳細をコード化できます。

For example, Mary Morris in the Manufacturing Division of Mitchell Cars in Detroit,
Michigan might have a digital certificate with a `SUBJECT` attribute of `C=US`,
`ST=Michigan`, `L=Detroit`, `O=Mitchell Cars`, `OU=Manufacturing`, `CN=Mary Morris /UID=123456`.
Mary's certificate is similar to her government identity card --- it provides
information about Mary which she can use to prove key facts about her. There are
many other attributes in an X.509 certificate, but let's concentrate on just these
for now.

たとえば、ミシガン州デトロイトのMitchell Cars製造部門のMary Morrisは、`SUBJECT`属性が`C = US`、`ST = Michigan`、`L = Detroit`、`O = Mitchell Cars`、`OU = Manufacturing`、`CN = Mary Morris /UID=123456`のデジタル証明書を持っている可能性があります 。
メアリーの証明書は政府の発行した身分証明書に似ています。メアリーに関する情報を提供して、彼女に関する重要な事実を証明できます。
X.509証明書には他にも多くの属性がありますが、ここではこれらだけに集中しましょう。

※訳注：属性の例についてはこちらのサイトを参照 https://certs.nii.ac.jp/archive/TSV_File_Format/csr/

![DigitalCertificate](./identity.diagram.8.png)

*A digital certificate describing a party called Mary Morris. Mary is the `SUBJECT` of the
certificate, and the highlighted `SUBJECT` text shows key facts about Mary. The
certificate also holds many more pieces of information, as you can see. Most importantly,
Mary's public key is distributed within her certificate, whereas her private signing key
is not. This signing key must be kept private.*

*メアリーモリスと呼ばれる団体を説明するデジタル証明書。
Maryは証明書の `SUBJECT(主題)` であり、強調表示された`SUBJECT`テキストはMaryに関する重要な事実を示しています。
ご覧のとおり、証明書にはさらに多くの情報が含まれています。
最も重要なことは、Maryの公開鍵は証明書内で配布されるのに対し、Maryの署名をした秘密鍵は配布されないことです。
この秘密鍵は非公開にする必要があります。*

What is important is that all of Mary's attributes can be recorded using a mathematical
technique called cryptography (literally, "*secret writing*") so that tampering will
invalidate the certificate. Cryptography allows Mary to present her certificate to others
to prove her identity so long as the other party trusts the certificate issuer, known
as a **Certificate Authority** (CA). As long as the CA keeps certain cryptographic
information securely (meaning, its own **private signing key**), anyone reading the
certificate can be sure that the information about Mary has not been tampered with ---
it will always have those particular attributes for Mary Morris. Think of Mary's X.509
certificate as a digital identity card that is impossible to change.

重要なのは、改ざんによって証明書が無効になるように、暗号化と呼ばれる数学的手法（文字通り「*シークレットライティング*」）を使用して、Maryのすべての属性を記録できることです。
暗号化により、Maryは、**認証局**（CA）として知られる証明書発行者を相手が信頼している限り、他人に証明書を提示して身元を証明することができます。
CAが特定の暗号化情報（つまり、**自身の署名に使った秘密鍵**）を安全に保持している限り、証明書を読んでいる人は誰でもMaryに関する情報が改ざんされていないことを確認できます。
メアリーのX.509証明書は、変更が不可能なデジタルIDカードだと考えてください。

## 認証、公開鍵、そして秘密鍵（Authentication, Public keys, and Private Keys）

Authentication and message integrity are important concepts in secure
communications. Authentication requires that parties who exchange messages
are assured of the identity that created a specific message. For a message to have
"integrity" means that cannot have been modified during its transmission.
For example, you might want to be sure you're communicating with the real Mary
Morris rather than an impersonator. Or if Mary has sent you a message, you might want
to be sure that it hasn't been tampered with by anyone else during transmission.

認証とメッセージの整合性は、安全な通信における重要な概念です。
認証では、メッセージを交換する関係者は、誰が特定のメッセージを作成したか保証される必要があります。
メッセージに「整合性」があるということは、その送信中に変更できなかったことを意味します。
たとえば、なりすましではなく、実際のメアリーモリスと通信していることを確認したい場合があります。
または、Maryからメッセージが送信された場合は、送信中に他の人によって改ざんされていないことを確認する必要があります。

Traditional authentication mechanisms rely on **digital signatures** that,
as the name suggests, allow a party to digitally **sign** its messages. Digital
signatures also provide guarantees on the integrity of the signed message.

従来の認証メカニズムは、名前が示すように、パーティがメッセージにデジタルに **署名できる** ようにする **デジタル署名** に依存しています。
デジタル署名は、署名されたメッセージの整合性も保証します。

Technically speaking, digital signature mechanisms require each party to
hold two cryptographically connected keys: a public key that is made widely available
and acts as authentication anchor, and a private key that is used to produce
**digital signatures** on messages. Recipients of digitally signed messages can verify
the origin and integrity of a received message by checking that the
attached signature is valid under the public key of the expected sender.

技術的に言えば、デジタル署名メカニズムでは、各当事者が2つの暗号接続キーを保持する必要があります。
広く利用可能とし、認証アンカーとして機能する公開鍵と、メッセージに **デジタル署名** を生成するために使用される秘密鍵です。
デジタル署名されたメッセージの受信者は、添付された署名が予想される送信者の公開キーの下で有効であることを確認することにより、受信したメッセージの発信元と整合性を検証できます。

**The unique relationship between a private key and the respective public key is the
cryptographic magic that makes secure communications possible**. The unique
mathematical relationship between the keys is such that the private key can be used to
produce a signature on a message that only the corresponding public key can match, and
only on the same message.

**秘密鍵とそれぞれの公開鍵の間のユニークな関係は、安全な通信を可能にする暗号魔法です。**
キー間の一意の数学的関係により、秘密キーを使用して、対応する公開キーのみが一致できるメッセージ、および同じメッセージのみに署名を作成できます。

![AuthenticationKeys](./identity.diagram.9.png)

In the example above, Mary uses her private key to sign the message. The signature
can be verified by anyone who sees the signed message using her public key.

上記の例では、Maryは秘密鍵を使用してメッセージに署名します。
署名は、公開鍵を使用して署名されたメッセージを見る人なら誰でも検証できます。

## 認証局（Certificate Authorities）

As you've seen, an actor or a node is able to participate in the blockchain network,
via the means of a **digital identity** issued for it by an authority trusted by the
system. In the most common case, digital identities (or simply **identities**) have
the form of cryptographically validated digital certificates that comply with X.509
standard and are issued by a Certificate Authority (CA).

これまで見てきたように、アクターまたはノードは、システムによって信頼された機関によって発行された **デジタルID** を使用して、ブロックチェーンネットワークに参加できます。
最も一般的なケースでは、デジタルID（または単に**ID**）は、X.509標準に準拠し、認証局（CA）によって発行された暗号で検証されたデジタル証明書の形式を持っています。

CAs are a common part of internet security protocols, and you've probably heard of
some of the more popular ones: Symantec (originally Verisign), GeoTrust, DigiCert,
GoDaddy, and Comodo, among others.

CAはインターネットセキュリティプロトコルの一般的な部分であり、Symantec（元はVerisign）、GeoTrust、DigiCert、GoDaddy、Comodoなど、より一般的なCAのいくつかを聞いたことがあるでしょう。

![CertificateAuthorities](./identity.diagram.11.png)

*A Certificate Authority dispenses certificates to different actors. These certificates
are digitally signed by the CA and bind together the actor with the actor's public key
(and optionally with a comprehensive list of properties). As a result, if one trusts
the CA (and knows its public key), it can trust that the specific actor is bound
to the public key included in the certificate, and owns the included attributes,
by validating the CA's signature on the actor's certificate.*

*認証局は、さまざまなアクターに証明書を配布します。
これらの証明書はCAによってデジタル署名され、アクターをそのアクターの公開鍵（およびオプションでプロパティの包括的なリスト）にバインドします。
その結果、CAを信頼する（およびその公開キーを知っている）場合、特定のアクターが証明書に含まれる公開鍵にバインドされ、アクターの証明書のCAの署名を検証することにより、含まれる属性を所有することを信頼できます。*

Certificates can be widely disseminated, as they do not include either the
actors' nor the CA's private keys. As such they can be used as anchor of
trusts for authenticating messages coming from different actors.

証明書には、アクターとCAの秘密キーのいずれも含まれていないため、広く配布できます。
そのため、異なるアクターからのメッセージを認証するための信頼のアンカーとして使用できます。

CAs also have a certificate, which they make widely available. This allows the
consumers of identities issued by a given CA to verify them by checking that the
certificate could only have been generated by the holder of the corresponding
private key (the CA).

CAにも証明書があり、広く利用可能になっています。
これにより、特定のCAによって発行されたIDのコンシューマーは、対応する秘密キー（CA）の所有者のみが証明書を生成できることを確認することで、それらを検証できます。

In a blockchain setting, every actor who wishes to interact with the network
needs an identity. In this setting, you might say that **one or more CAs** can be used
to **define the members of an organization's from a digital perspective**. It's
the CA that provides the basis for an organization's actors to have a verifiable
digital identity.

ブロックチェーン設定では、ネットワークとやり取りしたいすべてのアクターにアイデンティティが必要です。
この設定では、デジタルの観点から組織のメンバーを定義するために**1つ以上のCAを使用できる**と言えるかもしれません。
CAこそが、組織のアクターが、検証可能なデジタルIDを持つための基盤を提供しているからです。

### ルートCA、中間CA、信頼チェーン

CAs come in two flavors: **Root CAs** and **Intermediate CAs**. Because Root CAs
(Symantec, Geotrust, etc) have to **securely distribute** hundreds of millions
of certificates to internet users, it makes sense to spread this process out
across what are called *Intermediate CAs*. These Intermediate CAs have their
certificates issued by the root CA or another intermediate authority, allowing
the establishment of a "chain of trust" for any certificate that is issued by
any CA in the chain. This ability to track back to the Root CA not only allows
the function of CAs to scale while still providing security --- allowing
organizations that consume certificates to use Intermediate CAs with confidence
--- it limits the exposure of the Root CA, which, if compromised, would endanger
the entire chain of trust. If an Intermediate CA is compromised, on the other
hand, there will be a much smaller exposure.

CAには、**ルートCA**と**中間CA**の2つのフレーバーがあります。
ルートCA（Symantec、Geotrustなど）は、インターネットユーザーに何億もの証明書を**安全に配布**する必要があるため、このプロセスを*中間CA*と呼ばれるものに分散させることは理にかなっています。
これらの中間CAには、ルートCAまたは別の中間機関によって発行された証明書があり、チェーン内のCAによって発行された証明書の「信頼チェーン」を確立できます。
ルートCAに戻るこの機能により、CAの機能を拡張しながらセキュリティを確保できるだけでなく、証明書を使用する組織が自信を持って中間CAを使用できるようにするだけでなく、ルートCAの公開を制限します。 
ルートCAのセキュリティ侵害は、信頼チェーン全体を危険にさらします。
一方、中間CAが危険にさらされた場合は、危険にさらされる部分は、はるかに少なくなります。

![ChainOfTrust](./identity.diagram.1.png)

*A chain of trust is established between a Root CA and a set of Intermediate CAs
as long as the issuing CA for the certificate of each of these Intermediate CAs is
either the Root CA itself or has a chain of trust to the Root CA.*

*これらの各中間CAの証明書の発行CAがルートCA自体であるか、またはルートCAへの信頼チェーンがある限り、
ルートCAと一連の中間CAの間に信頼チェーンが確立されます。*

Intermediate CAs provide a huge amount of flexibility when it comes to the issuance
of certificates across multiple organizations, and that's very helpful in a
permissioned blockchain system (like Fabric). For example, you'll see that
different organizations may use different Root CAs, or the same Root CA with
different Intermediate CAs --- it really does depend on the needs of the network.

中間CAは、複数の組織にわたる証明書の発行に関して非常に大きな柔軟性を提供し、許可されたブロックチェーンシステム（Fabricなど）で非常に役立ちます。
たとえば、異なる組織が異なるルートCA、または異なる中間CAを持つ同じルートCAを使用する場合があります。実際には、ネットワークのニーズに依存します。


### Fabric CA

It's because CAs are so important that Fabric provides a built-in CA component to
allow you to create CAs in the blockchain networks you form. This component --- known
as **Fabric CA** is a private root CA provider capable of managing digital identities of
Fabric participants that have the form of X.509 certificates.
Because Fabric CA is a custom CA targeting the Root CA needs of Fabric,
it is inherently not capable of providing SSL certificates for general/automatic use
in browsers. However, because **some** CA must be used to manage identity
(even in a test environment), Fabric CA can be used to provide and manage
certificates. It is also possible --- and fully appropriate --- to use a
public/commerical root or intermediate CA to provide identification.

CAは非常に重要であるため、Fabricは組み込みのCAコンポーネントを提供して、形成するブロックチェーンネットワークにCAを作成できるようにします。
**Fabric CA**として知られるこのコンポーネントは、X.509証明書の形式を持つFabric参加者のデジタルIDを管理できるプライベートルートCAプロバイダーです。
Fabric CAは、FabricのルートCAのニーズを対象としたカスタムCAであるため、ブラウザで一般的/自動で使用するためのSSL証明書を本質的に提供することはできません。
ただし、**一部のCAは**（テスト環境であっても）IDの管理に使用する必要があるため、証明書を提供および管理するためにFabric CAを使用できます。
パブリック/コマーシャルルートまたは中間CAを使用して識別を提供することも可能です（そして、とても適切なやり方です）。

If you're interested, you can read a lot more about Fabric CA
[in the CA documentation section](http://hyperledger-fabric-ca.readthedocs.io/).

興味がある場合は、[CAのドキュメントセクション](http://hyperledger-fabric-ca.readthedocs.io/)でFabric CAの詳細を読むことができます。

## 証明書失効リスト

A Certificate Revocation List (CRL) is easy to understand --- it's just a list of
references to certificates that a CA knows to be revoked for one reason or another.
If you recall the store scenario, a CRL would be like a list of stolen credit cards.

証明書失効リスト（CRL）は簡単に理解できます。
これは、何らかの理由でCAが失効していることがわかっている証明書への参照のリストにすぎません。
食料品店でのシナリオを思い出すと、CRLは盗まれたクレジットカードのリストのようになります。

When a third party wants to verify another party's identity, it first checks the
issuing CA's CRL to make sure that the certificate has not been revoked. A
verifier doesn't have to check the CRL, but if they don't they run the risk of
accepting a compromised identity.

第三者が別の当事者の身元を確認する場合、最初に発行CAのCRLを確認して、証明書が取り消されていないことを確認します。
検証者はCRLをチェックする必要はありませんが、検証者は侵害されたIDを受け入れるリスクを負います。

![CRL](./identity.diagram.12.png)

*Using a CRL to check that a certificate is still valid. If an impersonator tries to
pass a compromised digital certificate to a validating party, it can be first
checked against the issuing CA's CRL to make sure it's not listed as no longer valid.*

*CRLを使用して、証明書がまだ有効であることを確認します。
なりすまし者が侵害されたデジタル証明書を検証側に渡そうとすると、最初に発行元のCAのCRLと照合して、有効でなくなったものとしてリストされていないことを確認できます。*

Note that a certificate being revoked is very different from a certificate expiring.
Revoked certificates have not expired --- they are, by every other measure, a fully
valid certificate. For more in-depth information about CRLs, click [here](https://hyperledger-fabric-ca.readthedocs.io/en/latest/users-guide.html#generating-a-crl-certificate-revocation-list).

失効する証明書は、有効期限切れの証明書とは大きく異なることに注意してください。
失効した証明書の有効期限は切れていません。他のすべての手段では、完全に有効な証明書です。
CRLの詳細については、[ここ](https://hyperledger-fabric-ca.readthedocs.io/en/latest/users-guide.html#generating-a-crl-certificate-revocation-list)をクリックしてください。

Now that you've seen how a PKI can provide verifiable identities through a chain of
trust, the next step is to see how these identities can be used to represent the
trusted members of a blockchain network. That's where a Membership Service Provider
(MSP) comes into play --- **it identifies the parties who are the members of a
given organization in the blockchain network**.

PKIが信頼チェーンを通じて検証可能なIDを提供する方法を確認したので、次のステップは、これらのIDを使用してブロックチェーンネットワークの信頼できるメンバーを表す方法を確認することです。
そこで、MSP（Membership Service Provider）が登場します。
**ブロックチェーンネットワーク内の特定の組織のメンバーである関係者を特定します。**

To learn more about membership, check out the conceptual documentation on [MSPs](../membership/membership.html).

メンバーシップの詳細については、[MSPに関する概念的なドキュメント](../membership/membership.html)をご覧ください。

<!---
Licensed under Creative Commons Attribution 4.0 International License https://creativecommons.org/licenses/by/4.0/
-->
