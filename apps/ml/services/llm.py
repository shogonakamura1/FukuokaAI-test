import os
from typing import List, Dict, Any
from openai import OpenAI

OPENAI_API_KEY = os.getenv("OPENAI_API_KEY", "")
OPENAI_MODEL = os.getenv("OPENAI_MODEL", "gpt-4o-mini")


class LLMService:
    def __init__(self):
        self.client = OpenAI(api_key=OPENAI_API_KEY) if OPENAI_API_KEY else None
        self.model = OPENAI_MODEL

    async def generate_recommendations(
        self,
        candidates: List[Dict[str, Any]],
        interest_tags: List[str],
        free_text: str = ""
    ) -> List[Dict[str, str]]:
        """候補スポットのおすすめ理由とレビュー要約を生成"""
        if not self.client:
            # APIキーがない場合はフォールバック
            return [
                {
                    "reason": "おすすめのスポットです",
                    "review_summary": "レビュー情報なし"
                }
                for _ in candidates
            ]

        # バッチで処理（コスト削減）
        reviews_text = []
        for candidate in candidates:
            reviews = candidate.get("reviews", [])
            if reviews:
                review_text = " ".join([r.get("text", "")[:200] for r in reviews[:3]])
                reviews_text.append(review_text)
            else:
                reviews_text.append("")

        # プロンプト作成
        prompt = f"""以下の福岡のスポットについて、ユーザーの興味（{', '.join(interest_tags)}）に基づいて、おすすめ理由（1〜2文）とレビュー要約（1〜2行）を生成してください。
{f'ユーザーの追加希望: {free_text}' if free_text else ''}

スポット情報:
"""
        for i, candidate in enumerate(candidates):
            prompt += f"""
{i+1}. {candidate.get('name', '')} ({candidate.get('category', '')})
レビュー: {reviews_text[i] if i < len(reviews_text) else ''}
"""

        prompt += """
JSON形式で返してください。各スポットについて以下の形式:
{
  "reason": "おすすめ理由（1〜2文）",
  "review_summary": "レビュー要約（1〜2行）"
}

配列形式で返してください。
"""

        try:
            response = self.client.chat.completions.create(
                model=self.model,
                messages=[
                    {"role": "system", "content": "あなたは福岡観光の専門家です。簡潔で魅力的な説明を生成してください。"},
                    {"role": "user", "content": prompt}
                ],
                temperature=0.7,
                max_tokens=1000,
            )

            content = response.choices[0].message.content
            # JSONをパース（簡易実装）
            import json
            # contentからJSON部分を抽出
            try:
                # JSON配列を探す
                start = content.find('[')
                end = content.rfind(']') + 1
                if start >= 0 and end > start:
                    json_str = content[start:end]
                    results = json.loads(json_str)
                    # 候補数に合わせて調整
                    while len(results) < len(candidates):
                        results.append({
                            "reason": "おすすめのスポットです",
                            "review_summary": "レビュー情報なし"
                        })
                    return results[:len(candidates)]
            except:
                pass

            # パース失敗時はフォールバック
            return [
                {
                    "reason": "おすすめのスポットです",
                    "review_summary": "レビュー情報なし"
                }
                for _ in candidates
            ]
        except Exception as e:
            print(f"LLM error: {e}")
            # エラー時はフォールバック
            return [
                {
                    "reason": "おすすめのスポットです",
                    "review_summary": "レビュー情報なし"
                }
                for _ in candidates
            ]


