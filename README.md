# p4s
p4s transfers files between PC and Switch. (yet another clone of Petit4Send in golang)

## (Original Petit4Send readme_ja.txt)
```
Petit4Send Ver. 1.2.2

概要
	PCからプチコン4へ、プチコン4からPCへファイルを転送するためソフトウェアです。
	PCからプチコン4への転送にはPro Micro(Arduino互換機)が2台、もしくはProMicro1台+UARTインタフェースが必要で、
	296Byte/Sec(2370bps)での転送が可能です。
	プチコン4からPCへは画像を使います。
	その他キーボードエミュレーションなどの機能があります。

動作環境
	PC
		.Net Framework 4
		USBポート
	Arduino互換機
		SparkFun Pro Micro 5V/16MHz 2台
			Arduino Uno等USBゲスト機能とハードウェアシリアルがある機種でも可能。
			一台はキーボードエミュレーション、一台はUART通信用。
			USB-UARTがある場合は1台でも可。
	Nintendo Switch
	プチコン4
	micro SDカード

仕様
	PC ←→ Pro Micro
		9600bps UART 双方向
	Pro Micro ←→ Switch
		USB HID Keyboard/Mouse併用
		最大2370bps(296Byte/sec)
		単方向
		エラー検出: CRC16(Koopman)
		圧縮: LZSS/無圧縮
	画像による転送
		スクリーンショットによる
		圧縮: LZSS/無圧縮
		181kByte/image(無圧縮時)

詳細は一次配布サイト(http://rei.to/petit4send.html)を参照ください。

インストール方法
	適当な場所に展開し、そのまま利用してください。

使用準備
	2台のProMicroはRT0とTXIをクロスするように接続し、さらにGNDを接続します。
	1台のProMicroをUSBでPCと接続し、[File]→[Write Firmware]から
	「Petit4Send.ino.promicro.hex」を書き込み、終わったら取り外します。
	こちらのProMicroがSwitch側になります。
	同様に、もう1台のProMicroに「USBUART.ino.promicro.hex」を書き込みます。
	こちらがPC側になります。

使用方法
	Switch側ProMicroをSwitchにつなぎ、PC側ProMicroをPCにつなぎます。
	プチコン4でPetit4Sendを起動します。
	Petit4Sendを起動し、ファイル、ファイルタイプを選びます。
	プチコン4側が[Wait]となってるのを確認し、
	PC側で[Send]を押します。
	プチコン4でファイル名、ファイルサイズ、ファイルタイプを確認し、待ちます。
	転送が完了したら「A」ボタンを押すと保存ダイアログがでるので保存します。
	
	詳細は一次配布サイト(http://rei.to/petit4send.html)を参照ください。

アンインストール方法
	レジストリなどは使わないのでフォルダごと削除します。

ライセンス・利用料・寄付
	このソフトは無料で利用して構いません。今後有料化の予定もありません。ご自由にお使いください。
	2次配布、商用利用、改造、リバースエンジニアリングなどもご自由にどうぞ。
	ソースコードは置いていませんが、欲しい人にはお渡しできます。
	また、寄付を受け付けています。詳細は一次配布サイト(http://rei.to/petit4send.html)を参照ください。

連絡先・著作者・一次配布サイト
	http://rei.to/petit4send.html
	Rei HOBARA reichan@white.plala.or.jp
	連絡をする場合、件名に「Petit4Send」の文字列が含まれるようにしてください。

ファイル構成
	Petit4Send.exe
		Petit4Sendの実行形式
	readme_ja.txt
		PetitModemPCの説明書き。このファイル。日本語。
	avrdude.exe
	avrdude.conf
	libusb0.dll
		ファームウェアを書き込むためのファイル。
		ArduinoIDEから抽出。
	Petit4Send.ino.promicro.hex
		Switch接続側のProMicro用のファームウェア。
	USBUART.ino.promicro.hex
		PC接続側のProMicro用のファームウェア。
	Petit4Sendフォルダ
		Petit4Send.ino.promicro.hexのソース
		コンパイルはファイル内の指示に従うこと
	USBUART
		USBUART.ino.promicro.hexのソース。
	Petit4Sendxxx.PRG
		Petit4Sendのプチコン4号用ソース。

注意事項
	http://rei.to/petit4send.html を参照してください。
```
