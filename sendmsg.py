import aiohttp

async def send_message(session, msg, phone_no):
    # URL of your Go server
    url = "http://localhost:8080/send"

    payload = {
        "jid": phone_no,
        "message": msg,
    }

    try:
        async with session.post(url, json=payload) as response:
            print("Status Code:", response.status)
            text = await response.text()
            print("Response Body:", text)
    except Exception as e:
        print(f"Send error: {e}")

if __name__ == "__main__":
    import asyncio

    async def main():
        async with aiohttp.ClientSession() as session:
            msg = "Hello, this is a test message!"
            phone_no = "918101858360"  # Replace with the actual phone number
            await send_message(session, msg, phone_no)

    asyncio.run(main())
# This script sends a message to a Go server running on localhost:8080