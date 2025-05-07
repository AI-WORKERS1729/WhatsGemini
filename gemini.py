import aiohttp
import asyncio
import os
import base64
from dotenv import load_dotenv
from sendMedia import send_media
from sendmsg import send_message

# Load environment variables from .env
load_dotenv()

GEMINI_API_KEY = os.getenv("GEMINI_API_KEY")

async def generate_reply(prompt_text, image_path='generated_images/gemini_image.png',phone_no=""):
    
    if prompt_text.startswith("generate a sticker"):
        # If the prompt is to generate a sticker, set the image path accordingly
        image_path = 'generated_images/gemini_sticker.webp'
        type = "sticker"
    else:
        # Otherwise, set the image path for normal image generation
        image_path = 'generated_images/gemini_image.png'
        type = "photo"
        
        
    url = f"https://generativelanguage.googleapis.com/v1beta/models/gemini-2.0-flash-exp-image-generation:generateContent?key={GEMINI_API_KEY}"

    headers = {
        "Content-Type": "application/json"
    }

    payload = {
        "contents": [{
            "parts": [{"text": prompt_text}]
        }],
        "generationConfig": {
            "responseModalities": ["TEXT", "IMAGE"]
        }
    }

    async with aiohttp.ClientSession() as session:
        try:
            async with session.post(url, headers=headers, json=payload) as response:
                if response.status == 200:
                    data = await response.json()
                    parts = data['candidates'][0]['content']['parts']

                    text_response = None
                    image_data = None

                    # Loop through parts to find text and image
                    for part in parts:
                        if 'text' in part:
                            text_response = part['text']
                        elif 'inlineData' in part and part['inlineData'].get('mimeType', '').startswith('image/'):
                            image_data = part['inlineData']['data']

                    if image_data:
                        os.makedirs(os.path.dirname(image_path), exist_ok=True)
                        with open(image_path, "wb") as img_file:
                            img_file.write(base64.b64decode(image_data))
                        if text_response:
                            await send_media(session, image_path, phone_no, type, text_response)
                        else:
                            await send_media(session, image_path, phone_no, type, "Check this out!")
                            # print("Image generated and saved.")
                    elif text_response:
                        await send_message(session, text_response, phone_no)
                        # print("Text Response:", text_response)
                    else:
                        print("No text or image content found in the response.")
                        return "No content generated."

                    # # Return logic
                    # if text_response and image_data:
                    #     return text_response  # both present, return text
                    # elif text_response:
                    #     return text_response  # only text
                    # elif image_data:
                    #     return None  # only image
                    # else:
                    #     return "No content generated."

                else:
                    print(f"Gemini error: {response.status}")
                    return "Sorry, I'm unable to respond right now."

        except Exception as e:
            print(f"Gemini exception: {e}")
            return "Sorry, I'm unable to respond right now."



if __name__ == "__main__":
    # Example usage
    async def main():
        prompt = "Create a beautiful sunset over the ocean."
        response = await generate_reply(prompt)
        async with aiohttp.ClientSession() as session:
            phone_no = "918101858360"
            if response:
                await send_message(session, response, phone_no)
                # print("Text Response:", response)
            else:
        
                # Send the generated image
                await send_media(session, "generated_images/gemini_image.png", phone_no, "photo", "Check this out!")
                print("Image generated and saved.")
        # print(response)

    asyncio.run(main())