import asyncio
import aiohttp
from fetchmsg import fetch_message
from sendmsg import send_message
from gemini import generate_reply


default_prompt="Respond to the following message in a friendly tone with emojis and a relaxed vibe, just like a real human would in the same language as the user asked. The response should be short, cheerful, and engaging without extra details. If image generation is requested,  generate the image and provide a brief description of the image max 20 words also"
async def message_worker():
    async with aiohttp.ClientSession() as session:
        while True:
            fetched_message = await fetch_message(session)

            
            if fetched_message:
                print(fetched_message)
                input_text = fetched_message["message"]
                if "[Sticker message]"==input_text:
                    await generate_reply(f"only generate a sticker don't generate any metadata or any text",phone_no=fetched_message["from"])
                else:
                    await generate_reply(f"{default_prompt} {input_text}",phone_no=fetched_message["from"])
                # await send_message(session, output_text, fetched_message["from"])
            else:
                await asyncio.sleep(2.5)  # Sleep if no new messages

async def main():
    print("📨 Async WhatsApp Bot started. Waiting for incoming messages...")
    await message_worker()

if __name__ == "__main__":
    try:
        asyncio.run(main())
    except KeyboardInterrupt:
        print("❌ Bot stopped by user.")

    