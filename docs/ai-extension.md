# Slink AI 图片处理外置扩展

智能去水印、AI 超分辨率（高清放大）、智能识别打标等能力依赖真实的机器学习模型，
**无法在 Slink 的纯 Go 进程内离线完成**。因此 Slink 把这些能力设计为「外置扩展」：
主程序只负责鉴权、参数转发与结果回传，真正的推理交给一个**可自行部署、可替换**的外部服务。

> 仅需「压缩 / 格式转换 / 高质量缩放放大 / 缩略图」？这些是纯本地能力，无需任何外部服务，
> 直接用 `POST /api/image/process`（见下文「本地 vs 外置」）。

---

## 1. 能力一览

| 能力 | capability | 路由 | 返回 |
| --- | --- | --- | --- |
| 智能去水印 | `dewatermark` | `POST /api/image/ai/dewatermark` | 处理后图片 |
| AI 超分辨率/高清放大 | `upscale` | `POST /api/image/ai/upscale` | 处理后图片 |
| 智能识别/自动打标 | `tag` | `POST /api/image/ai/tag` | 标签数组 |
| 扩展状态 | — | `GET /api/image/ai/status` | 是否已配置 |

所有 AI 路由都需要登录（JWT 或 Bearer Token），与图片上传接口一致。

**未配置外部服务时**，去水印/超分/打标接口返回 `501 Not Implemented`，并在响应体中
说明这是外置扩展、如何启用。这是预期行为，便于前端优雅降级、隐藏入口。

---

## 2. 启用方式

有两种方式配置，二选一即可：

### 方式一：后台「AI 设置」页（推荐，可视化）

管理员登录后进入 **系统 → AI 设置**，填写：

- **图像处理外置扩展**：服务地址 Endpoint、令牌 Token、超时
- **LLM 平台**（OpenAI 兼容）：Base URL、API Key、默认模型
- **向量 Embedding 平台**（OpenAI 兼容）：Base URL、API Key、默认模型、维度

每块都有「测试连接」按钮。配置入库（`configs` 表），密钥在接口返回时脱敏展示，
保存时留空表示保留原值。LLM / Embedding 当前用于配置与连通性测试，能力预留。

### 方式二：环境变量（优先级更高，适合 Docker 注入密钥）

设置了下列环境变量的字段会**覆盖**后台配置，并在后台显示为只读（避免明文落库）：

| 变量 | 说明 |
| --- | --- |
| `SLINK_AI_ENDPOINT` | 图像扩展服务地址，如 `http://127.0.0.1:9000`。未配置即视为未启用。 |
| `SLINK_AI_TOKEN` | 图像扩展令牌，作为 `Authorization: Bearer <token>` 头发送。 |
| `SLINK_AI_TIMEOUT` | 图像扩展请求超时秒数，默认 `120`。 |
| `SLINK_AI_LLM_BASE_URL` / `SLINK_AI_LLM_API_KEY` / `SLINK_AI_LLM_MODEL` | LLM 平台 |
| `SLINK_AI_EMBEDDING_BASE_URL` / `SLINK_AI_EMBEDDING_API_KEY` / `SLINK_AI_EMBEDDING_MODEL` / `SLINK_AI_EMBEDDING_DIMENSIONS` | Embedding 平台 |

docker-compose 示例：

```yaml
services:
  slink:
    environment:
      SLINK_AI_ENDPOINT: http://ai-ext:9000
      SLINK_AI_TOKEN: ${SLINK_AI_TOKEN:-}
      SLINK_AI_TIMEOUT: "120"
```

配置后无需改动主程序，重启即可。是否生效可调用 `GET /api/image/ai/status` 查看
`configured: true`。

---

## 3. 外部服务需实现的契约

Slink 以 provider 无关的 HTTP/JSON 契约调用外部服务。你可以用任意语言/框架实现它，
只要满足下述约定（可基于 lama-cleaner、Real-ESRGAN、自建推理服务或第三方 SaaS 封装）。

### 请求

```
POST  {SLINK_AI_ENDPOINT}/v1/{capability}
Header: Content-Type: application/json
        Authorization: Bearer <SLINK_AI_TOKEN>   # 若配置了 token
Body:
{
  "image": "<base64 编码的原图字节>",
  "mime_type": "image/png",
  "params": {                 // 可选，按能力透传
    "scale": 2,               // upscale：放大倍数
    "regions": [[x,y,w,h]],   // dewatermark：水印区域（可选）
    "prompt": "...",          // 可选：文本提示
    "model": "..."            // 可选：指定模型
  }
}
```

### 响应

图像类能力（`dewatermark` / `upscale`），返回 `200`：

```json
{
  "image": "<base64 处理后图片>",
  "mime_type": "image/png",
  "meta": { "any": "可选透传信息" }
}
```

识别类能力（`tag`），返回 `200`：

```json
{
  "tags": [
    { "name": "cat", "score": 0.97 },
    { "name": "outdoor", "score": 0.88 }
  ],
  "meta": {}
}
```

错误时返回非 `2xx`，并在体中带 `error`：

```json
{ "error": "模型加载失败" }
```

Slink 会将该 `error` 文案透传给调用方（HTTP `502`）。响应体上限 64 MiB。

---

## 4. 参考实现骨架（Python / FastAPI）

下面是一个最小可替换的外部服务骨架，把推理部分换成你选用的模型即可：

```python
from fastapi import FastAPI, Header, HTTPException
from pydantic import BaseModel
import base64

app = FastAPI()

class Req(BaseModel):
    image: str
    mime_type: str | None = None
    params: dict | None = None

def check(auth: str | None):
    # 如启用了 SLINK_AI_TOKEN，在此校验
    pass

@app.post("/v1/upscale")
def upscale(req: Req, authorization: str | None = Header(None)):
    check(authorization)
    raw = base64.b64decode(req.image)
    scale = (req.params or {}).get("scale", 2)
    out = run_real_esrgan(raw, scale)          # TODO: 接入 Real-ESRGAN
    return {"image": base64.b64encode(out).decode(), "mime_type": "image/png"}

@app.post("/v1/dewatermark")
def dewatermark(req: Req, authorization: str | None = Header(None)):
    check(authorization)
    raw = base64.b64decode(req.image)
    out = run_lama_cleaner(raw, (req.params or {}).get("regions"))  # TODO: 接入 lama-cleaner
    return {"image": base64.b64encode(out).decode(), "mime_type": "image/png"}

@app.post("/v1/tag")
def tag(req: Req, authorization: str | None = Header(None)):
    check(authorization)
    raw = base64.b64decode(req.image)
    tags = run_tagger(raw)                      # TODO: 接入打标模型
    return {"tags": [{"name": n, "score": s} for n, s in tags]}
```

可选用的开源模型/服务参考：

- 去水印 / 图像修复：[lama-cleaner](https://github.com/Sanster/lama-cleaner)、IOPaint、LaMa
- 超分辨率：[Real-ESRGAN](https://github.com/xinntao/Real-ESRGAN)、GFPGAN、SwinIR
- 智能打标：CLIP、wd-tagger、各类图像分类/多标签模型

---

## 5. 本地能力 vs 外置扩展

| | 本地图片处理 | AI 外置扩展 |
| --- | --- | --- |
| 路由 | `/api/image/process` | `/api/image/ai/*` |
| 依赖 | 无（纯 Go，进程内完成） | 外部 AI 服务 |
| 能力 | 压缩、格式转换、高质量缩放/放大、缩略图 | 智能去水印、AI 超分、智能打标 |
| 高清放大 | 传统高质量重采样（CatmullRom 近似双三次） | AI 超分（细节生成） |
| 离线可用 | 是 | 否（需配置 `SLINK_AI_ENDPOINT`） |

两者互补：例如可先用本地 `process` 做格式转换/压缩，再调用外置 `upscale` 做 AI 高清放大。

---

## 6. 调用示例

```bash
# 查看扩展状态
curl -H "Authorization: Bearer <TOKEN>" \
  http://localhost:8080/api/image/ai/status

# AI 超分（返回二进制图片，放大 2 倍）
curl -H "Authorization: Bearer <TOKEN>" \
  -F image=@photo.jpg -F scale=2 \
  http://localhost:8080/api/image/ai/upscale -o photo_2x.png

# 智能去水印（返回 base64 JSON）
curl -H "Authorization: Bearer <TOKEN>" \
  -F image=@photo.jpg -F response=json \
  http://localhost:8080/api/image/ai/dewatermark

# 智能打标
curl -H "Authorization: Bearer <TOKEN>" \
  -F image=@photo.jpg \
  http://localhost:8080/api/image/ai/tag
```
