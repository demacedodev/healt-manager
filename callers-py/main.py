import tinytuya
from ddtrace import patch_all
from fastapi import FastAPI, HTTPException
from fastapi.middleware.cors import CORSMiddleware

patch_all()

app = FastAPI()

app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)


@app.get("/ping")
def ping():
    return "pong"


@app.get("/health/device/{device_id}/status")
async def get_device_status(device_id: str, device_ip: str, device_password: str):
    try:
        return device_status(device_id, device_ip, device_password)
    except ServiceError as e:
        raise HTTPException(status_code=e.code, detail=e.message)


def device_status(device_id: str, device_ip: str, device_password: str) -> dict:
    device = tinytuya.OutletDevice(dev_id=device_id, address=device_ip, local_key=device_password, version=3.4)
    status = device.status()

    dps = status.get('dps')
    if dps is None:
        return {
            "device_ip": device_ip,
            "device_id": device_id,
            "device_status": False,
            "raw_response": status,
            "error": {
                "code": "ERR-001",
                "message": "has no dps information maybe device is down",
            }
        }

    dps_status = dps.get('1')
    if dps_status is None:
        return {
            "device_ip": device_ip,
            "device_id": device_id,
            "device_status": False,
            "raw_response": status,
            "error": {
                "code": "ERR-002",
                "message": "has no status information",
            }
        }

    return {
        "device_ip": device_ip,
        "device_id": device_id,
        "device_status": dps_status,
        "raw_response": status,
        "error": None
    }


class ServiceError(Exception):
    def __init__(self, code: int, message: str):
        self.code = code
        self.message = message
        super().__init__(self.code, self.message)

# if __name__ == "__main__":
#     uvicorn.run("main:app", host="0.0.0.0", port=8081, log_level="info", reload=True)
