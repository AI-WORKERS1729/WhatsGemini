import aiohttp

async def fetch_message(session):
    # URL of your Go server
    url = "http://localhost:8080/messages"

    try:
        async with session.get(url) as response:
            if response.status == 200:
                messages = await response.json()
                if messages:
                    print("📩 Received Messages:")
                    return messages[-1]  # Return the last message only
                else:
                    return None
            else:
                print("Failed to fetch messages:", response.status)
                return None
    except Exception as e:
        print(f"Fetch error: {e}")
        return None
    
if __name__ == "__main__":
    import asyncio

    async def main():
        async with aiohttp.ClientSession() as session:
            message = await fetch_message(session)
            if message:
                print("Last message:", message)
            else:
                print("No messages received.")

    asyncio.run(main())
# This script fetches messages from a Go server running on localhost:8080