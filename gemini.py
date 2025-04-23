import aiohttp
import os
from dotenv import load_dotenv

# Load environment variables from .env
load_dotenv()

GEMINI_API_KEY = os.getenv("GEMINI_API_KEY")  # 🔥 Best practice: read API key from env variable

async def generate_reply(prompt_text):
    url = f"https://generativelanguage.googleapis.com/v1beta/models/gemini-2.0-flash:generateContent?key={GEMINI_API_KEY}"
    headers = {
        "Content-Type": "application/json"
    }
    payload = {
        "contents": [{
            "parts": [{"text": prompt_text}]
        }]
    }

    async with aiohttp.ClientSession() as session:
        try:
            async with session.post(url, headers=headers, json=payload) as response:
                if response.status == 200:
                    data = await response.json()
                    # Extract generated text
                    text = data['candidates'][0]['content']['parts'][0]['text']
                    return text
                else:
                    print(f"Gemini error: {response.status}")
                    return "Sorry, I'm unable to respond right now."
        except Exception as e:
            print(f"Gemini error: {e}")
            return "Sorry, I'm unable to respond right now."
