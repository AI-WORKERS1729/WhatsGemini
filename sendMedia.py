import os
import aiohttp

async def send_media(session, file_path, phone_no, media_type, caption=None):
    # URL of your Go server
    url = "http://localhost:8080/sendMedia"
    
    
    # Ensure to use absolute file path
    payload = {
        "jid": phone_no,
        "file_path": os.path.abspath(file_path),
        "media_type": media_type,
        "caption": caption,
    }
    
    # if caption and media_type in ["image/jpeg", "video/mp4"]:  # Check if caption is needed
    #     payload["caption"] = caption
    
    try:
        async with session.post(url, json=payload) as response:
            print("Status Code:", response.status)
            text = await response.text()
            
            if response.status == 200:
                print(f"{media_type.capitalize()} Response:", text)
            else:
                print(f"Failed to send {media_type}. Response: {text}")
    except Exception as e:
        print(f"Send Media error: {e}")



# async def main():
#     async with aiohttp.ClientSession() as session:
#         phone_no = "918101858360"  # Replace with a valid JID

#         # Send a text message
#         msg = "Hello from the updated Python client!"
#         await send_message(session, msg, phone_no)

#         # Send media
#         # await send_media(session, "tnienow.jpeg", phone_no, "photo", "Check this out!")
#         # await send_media(session, "/absolute/path/to/video.mp4", phone_no, "video", "Watch this video!")
#         # await send_media(session, "/absolute/path/to/audio.ogg", phone_no, "audio")
#         # await send_media(session, "/absolute/path/to/file.pdf", phone_no, "document")
#         await send_media(session, "tnienow.webp", phone_no, "sticker")

#         # Fetch received messages
#         # await fetch_messages(session)

# # Run the main function asynchronously
# asyncio.run(main())
