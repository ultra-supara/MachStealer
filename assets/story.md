# The Story Behind the Banner

> Visual identity for **MachStealer: One Pipeline Behind Every macOS Infostealer**
> Black Hat USA 2026 Arsenal — Mandalay Bay, Las Vegas
> Files: `machstealer-banner.svg`, `machstealer-mark.svg`, `machstealer-favicon.svg`

---

### Core message

> **This is not a story about how data is stolen. It is a story about the silence that remains after the theft.**

After an infostealer runs, the victim's Mac keeps behaving normally. The fans don't spin up. The screen doesn't change. The Keychain prompt closes. That *unnoticed-ness* is the essence of macOS infostealers — and the banner translates it into the visual language of a Japanese garden at night.

### Motif by motif

#### 1. *Izayoi* — the moon on the 16th night

> "This is *izayoi* — the moon on the 16th night of the lunar cycle. One night past full. In Japanese aesthetics, *izayoi* is considered more beautiful than the full moon, because **imperfection invites contemplation**.
>
> Infostealers are the same. They never give you a complete picture — a few cookies missing, a Keychain prompt you barely remember dismissing, a process that exited cleanly. The picture is *almost* whole. That gap is where the harm lives."

**Meaning:** The moon represents the trace the victim almost noticed, but didn't.

#### 2. The keyhole inside the moon

> "Look at the moon again. There's a keyhole at its center. That is *Keychain*. Every macOS infostealer family — AMOS, Poseidon, Banshee, Cthulhu, Cuckoo — they all start by knocking on this single door.
>
> And the door always answers. `security find-generic-password -wa Chrome` returns the master key. PBKDF2, 1003 iterations, salt 'saltysalt' — those constants haven't changed in a decade. The moon is red because the door has been opened."

**Meaning:** There is only one keyhole. The attack surface is singular, and that is where defense must concentrate.

#### 3. The vermilion — warning, and signature

> "The vermilion is not a Japanese flag. It is the color of a *rakkan* — the seal an artist presses onto a finished work to claim authorship. In the lower corner you see one: 解体録, *kaitai-roku* — 'dissection record.' That is what this tool is. Not malware. A signed, dated dissection.
>
> The same vermilion is on the moon, because the act of stealing also leaves a signature — if you know how to read it. My talk is about teaching defenders to read that signature."

**Meaning:** Red marks both the attacker's trace and the researcher's accountability. Both are *signed*.

#### 4. The pine branch

> "The pine in Japanese gardens is *tokowa* — evergreen, the symbol of what does not change. In our story, the pine is the security researcher. We have been watching the same pipeline — Keychain, PBKDF2, AES-128-CBC, SQLite copy — for ten years. The malware authors change names, change C2 servers, rebrand from AMOS to Poseidon to Banshee. The pine does not move."

**Meaning:** Malware is transient; the defender community is continuous.

#### 5. *Karesansui* — the dry garden, the water that does not flow

> "*Karesansui* is a dry garden. There is no water — only sand, raked into ripples that *suggest* water. It is a meditation on what isn't there.
>
> MachStealer is a karesansui. There is no exfiltration. No C2. No persistence. Just the *shape* of an infostealer, raked into disk, so you can see it without ever running real malware in your environment. The three stones at the bottom — *sangen-iwagumi*, the three-stone arrangement — represent the three things every family steals: cookies, passwords, cards."

**Meaning:** *Karesansui* is the deliberate omission. The PoC ethics — what was intentionally left out — are visualized.

#### 6. Vertical kanji "侘寂の罠"

> "*Wabi-sabi* — the Japanese acceptance of imperfection, transience, and the beauty in things that decay. The kanji on the right reads: *wabi-sabi no wana* — the **trap** of wabi-sabi.
>
> Because that is what makes infostealers so effective on macOS. The OS feels quiet, polished, *finished*. Users have been trained to interpret silence as safety. But silence is also what an infostealer leaves behind. The same aesthetic that makes a Mac feel calm is what lets the theft go unnoticed."

**Meaning:** The minimalist aesthetic of macOS becomes the attacker's cover.

#### 7. Negative space — the most important element

> "Notice how much of this image is *empty*. In Japanese aesthetics, the negative space — *ma* — is not absence. It is where meaning lives.
>
> What MachStealer does *not* do is more important than what it does. No network. No persistence. No evasion. The empty space in this banner is the part of an infostealer I deliberately did not build. That is the talk."

**Meaning:** The empty space is the ethical statement. The code I did not write is the strongest message.

### Anticipated audience questions

| Question | Reply |
|---|---|
| Why a Japanese garden? | "Because macOS aesthetic and Japanese aesthetic share a value — *the calm finished surface*. That is also what makes infostealers work." |
| Is the moon Apple? | "No. The moon is the user's machine after the theft — looks whole, isn't." |
| What does the seal say? | "*Kaitai-roku* — dissection record. It is a research artifact, signed." |
| Why no exfil in the tool? | (Point to the seal in the lower-right.) "Because this is signed as a dissection, not a weapon." |
---

### 物語の核

> 「これは情報を盗む技術の話ではない。盗まれた後に残る "静けさ" の話だ。」

侵入後、被害者の Mac はいつも通りに動き続ける。ファンは唸らず、画面は変わらず、Keychain のプロンプトは閉じる。その**気づかれなさ**こそが macOS infostealer の本質で、それを夜の日本庭園の言語で翻訳したものがこのバナーである。

### モチーフごとの語り

#### 1. 十六夜の月 — 不完全の美

> 「これは十六夜 — 旧暦十六夜の月。満月から一晩過ぎた月だ。日本の美意識では、満月よりも十六夜の方が美しいとされる。不完全さこそが、観る者に思考を促すからだ。
>
> Infostealer も同じだ。決して完全な像を残さない。クッキーが幾つか欠けている。Keychain のプロンプトを閉じた記憶が朧げにある。プロセスは綺麗に終了している。像は ほぼ 完成している。その隙間に害が住んでいる。」

含意: 月は、被害者が「気づきかけて気づけなかった」痕跡そのもの。

#### 2. 月の中の鍵穴 — Keychain

> 「月をもう一度見てほしい。中央に鍵穴がある。これが Keychain だ。AMOS、Poseidon、Banshee、Cthulhu、Cuckoo — どの macOS infostealer ファミリーも、まずこのひとつの扉を叩くところから始める。
>
> そして扉は必ず応える。`security find-generic-password -wa Chrome` がマスターキーを返す。PBKDF2、1003 iterations、salt は 'saltysalt' — この定数は十年変わっていない。月が朱いのは、扉が開かれたからだ。」

含意: 鍵穴はひとつしかない = 攻撃面の一意性。ここが防御の急所。

#### 3. 朱の色

> 「この朱は落款 — 作家が完成した作品に押す印の朱だ。右下に小さく押されている：『解体録』。これがこのツールの正体だ。マルウェアではない。署名され、日付の入った解体記録だ。
>
> 同じ朱が月にもある。盗む行為もまた署名を残すからだ — 読み方さえ知っていれば。私の講演はその署名の読み方を防御者に渡すことだ。」

含意: 朱は攻撃の痕跡であり、研究者の責任でもある。どちらも "署名されている" のが本質。

#### 4. 松の枝

> 「日本庭園の松は常磐 — 常緑、変わらぬものの象徴だ。この物語において、松は研究者である。我々はこの十年、同じパイプラインを観察してきた — Keychain, PBKDF2, AES-128-CBC, SQLite コピー。マルウェア作者は名前を変える、C2 を変える、AMOS から Poseidon へ、Banshee へ rebrand する。松は動かない。」

含意: マルウェアは流転する。防御コミュニティは継続する。

#### 5. 枯山水

> 「枯山水 は水の無い庭だ。砂を熊手で掻いて、水であるかのような波紋を作る。そこに無いもの に対する瞑想なのだ。
> MachStealer は枯山水だ。Exfiltration なし。C2 なし。Persistence なし。Infostealer の形だけが、ディスクに掻かれた波紋として残る。本物のマルウェアを実環境で動かすことなく、その姿を観察できる。底にある三つの石は三尊石組 — どの infostealer ファミリーも盗む三つのもの (cookies / passwords / cards) を表している。」

含意: 枯山水は意図的な省略の表現。PoC として削ぎ落とした倫理を視覚化している。

#### 6. 縦書き「侘寂の罠」

> 「侘寂 — 不完全さ、無常、朽ちていくものに宿る美を受け入れる、日本の美意識。右側の漢字を読むと『侘寂の罠』となる。
>
> なぜならそれが、macOS において infostealer が効果的である理由だからだ。OS は静かで、磨かれていて、完成された ものとして体験される。ユーザーは静寂を安全と解釈するよう訓練されてきた。しかし静寂はまた、infostealer が後に残すものでもある。Mac を穏やかに感じさせる美意識そのものが、窃取を不可視にしている。」

含意: macOS のミニマリズム = 攻撃者の隠蔽装置という主題。

#### 7. 余白

> 「この画像のどれだけが空白か見てほしい。日本の美意識において、空白、 間は不在ではない。意味が宿る場所だ。
>
> MachStealer が しないことは、する ことよりも重要だ。ネットワークなし。Persistence なし。Evasion なし。このバナーの空白は、私が意図的に作らなかった infostealer の部分だ。それがこの講演の主題だ。」

含意: 余白は倫理の表明。書かなかったコードが一番のメッセージ。

### 想定 Q&A

| 想定質問 | 答え方 |
|---|---|
| なぜ日本庭園？ | 「macOS の美意識と日本の美意識はひとつの価値を共有している — **静かに完成された表面**。それは infostealer が機能する理由でもある。」 |
| 月は Apple のロゴ？ | 「違う。月は窃取後の被害者の機械だ — 完全に見えるが、完全ではない。」 |
| 落款の文字は？ | 「**解体録** — dissection record。これが研究成果物であることを署名している。」 |
| なぜツールに exfiltration が無い？ | (右下の落款を指して) 「この作品は武器ではなく解体記録として署名されているからだ。」 |
