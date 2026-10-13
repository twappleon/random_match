from __future__ import annotations

from pathlib import Path
import math

from PIL import Image, ImageDraw, ImageFilter, ImageFont


ROOT = Path(__file__).resolve().parents[1]
OUT = ROOT / "assets" / "google-play"
SCREENSHOT_OUT = OUT / "phone-screenshots"
LOGO = ROOT / "assets" / "brand" / "logo-options" / "option-2.png"

W, H = 1080, 1920

FONT_CN = "/System/Library/Fonts/STHeiti Medium.ttc"
FONT_CN_LIGHT = "/System/Library/Fonts/STHeiti Light.ttc"
FONT_LATIN = "/System/Library/Fonts/HelveticaNeue.ttc"


def font(size: int, bold: bool = False) -> ImageFont.FreeTypeFont:
    if bold and Path(FONT_CN).exists():
        path = FONT_CN
    elif Path(FONT_CN_LIGHT).exists():
        path = FONT_CN_LIGHT
    else:
        path = FONT_CN if Path(FONT_CN).exists() else FONT_LATIN
    return ImageFont.truetype(path, size=size, index=0)


def rounded(draw: ImageDraw.ImageDraw, box, radius, fill, outline=None, width=1):
    draw.rounded_rectangle(box, radius=radius, fill=fill, outline=outline, width=width)


def shadow_layer(size, box, radius, blur=28, alpha=70):
    layer = Image.new("RGBA", size, (0, 0, 0, 0))
    d = ImageDraw.Draw(layer)
    d.rounded_rectangle(box, radius=radius, fill=(0, 0, 0, alpha))
    return layer.filter(ImageFilter.GaussianBlur(blur))


def gradient(size, top, bottom):
    w, h = size
    img = Image.new("RGB", size, top)
    px = img.load()
    for y in range(h):
        t = y / max(1, h - 1)
        for x in range(w):
            dt = min(1, max(0, t + x / w * 0.14))
            px[x, y] = tuple(int(top[i] * (1 - dt) + bottom[i] * dt) for i in range(3))
    return img.convert("RGBA")


def add_brand_bg(img, accent=(76, 220, 215), warm=(255, 85, 146)):
    layer = Image.new("RGBA", img.size, (0, 0, 0, 0))
    d = ImageDraw.Draw(layer)
    d.ellipse((-190, 130, 285, 605), fill=(*accent, 26))
    d.ellipse((780, 120, 1240, 580), fill=(*warm, 22))
    d.ellipse((720, 1390, 1170, 1840), fill=(*accent, 16))
    for i in range(8):
        x = 120 + i * 130
        y = 1320 + int(math.sin(i) * 42)
        d.line((x, y, x + 220, y - 180), fill=(255, 255, 255, 18), width=3)
    img.alpha_composite(layer.filter(ImageFilter.GaussianBlur(1)))


def text(draw, xy, s, size, fill=(255, 255, 255), bold=False, spacing=8, anchor=None, align="left"):
    draw.multiline_text(
        xy,
        s,
        font=font(size, bold),
        fill=fill,
        spacing=spacing,
        anchor=anchor,
        align=align,
    )


def centered_text(draw, box, s, size, fill, bold=False):
    f = font(size, bold)
    bbox = draw.multiline_textbbox((0, 0), s, font=f, spacing=6, align="center")
    tw, th = bbox[2] - bbox[0], bbox[3] - bbox[1]
    x = box[0] + (box[2] - box[0] - tw) / 2
    y = box[1] + (box[3] - box[1] - th) / 2
    draw.multiline_text((x, y), s, font=f, fill=fill, spacing=6, align="center")


def pill(draw, x, y, label, fill, fg=(255, 255, 255), pad_x=22, h=46):
    f = font(22)
    bbox = draw.textbbox((0, 0), label, font=f)
    w = bbox[2] - bbox[0] + pad_x * 2
    rounded(draw, (x, y, x + w, y + h), h // 2, fill)
    draw.text((x + pad_x, y + 10), label, font=f, fill=fg)
    return x + w


def app_frame(img, x=70, y=425, w=940, h=1360):
    img.alpha_composite(shadow_layer(img.size, (x, y, x + w, y + h), 44, 30, 70))
    d = ImageDraw.Draw(img)
    rounded(d, (x, y, x + w, y + h), 44, (10, 15, 29), (72, 88, 120), 1)
    rounded(d, (x + 18, y + 18, x + w - 18, y + h - 18), 34, (11, 16, 31))
    d.text((x + 48, y + 36), "9:41", font=font(22, True), fill=(218, 228, 248))
    sx = x + w - 148
    d.ellipse((sx, y + 42, sx + 16, y + 58), fill=(168, 181, 210))
    d.rounded_rectangle((sx + 34, y + 40, sx + 46, y + 60), 3, outline=(168, 181, 210), width=2)
    d.rectangle((sx + 37, y + 55, sx + 43, y + 58), fill=(168, 181, 210))
    d.rounded_rectangle((sx + 68, y + 40, sx + 82, y + 60), 3, outline=(168, 181, 210), width=2)
    d.rectangle((sx + 71, y + 52, sx + 79, y + 58), fill=(168, 181, 210))
    return (x + 54, y + 92, x + w - 54, y + h - 38)


def header(img, title, subtitle, eyebrow="Interest Match Online"):
    d = ImageDraw.Draw(img)
    text(d, (70, 70), eyebrow, 27, fill=(103, 232, 224), bold=True)
    text(d, (70, 126), title, 58, fill=(255, 255, 255), bold=True, spacing=10)
    text(d, (72, 270), subtitle, 27, fill=(213, 224, 246), spacing=8)


def draw_nav(d, box, active):
    x1, y1, x2, y2 = box
    nav = (x1 + 2, y2 - 112, x2 - 2, y2 - 14)
    rounded(d, nav, 28, (23, 30, 48), (50, 63, 91), 1)
    labels = ["连线", "探索", "资料", "安全"]
    sw = (nav[2] - nav[0]) / 4
    for i, lab in enumerate(labels):
        xx = nav[0] + i * sw
        if i == active:
            rounded(d, (xx + 10, nav[1] + 10, xx + sw - 10, nav[3] - 10), 22, (106, 88, 255))
            fill = (255, 255, 255)
        else:
            fill = (151, 164, 195)
        centered_text(d, (xx, nav[1] + 10, xx + sw, nav[3] - 12), lab, 23, fill, i == active)


def draw_video_screen(d, box):
    x1, y1, x2, y2 = box
    d.text((x1 + 8, y1 + 2), "兴趣匹配", font=font(34, True), fill=(255, 255, 255))
    for i, (num, lab) in enumerate([("12,480", "在线"), ("318", "等待"), ("2,106", "聊天")]):
        x = x1 + 12 + i * ((x2 - x1 - 24) / 3)
        rounded(d, (x, y1 + 68, x + 245, y1 + 152), 18, (26, 34, 55), (60, 76, 109), 1)
        d.text((x + 18, y1 + 82), num, font=font(29, True), fill=(255, 255, 255))
        d.text((x + 18, y1 + 118), lab, font=font(18), fill=(148, 160, 188))
    card = (x1 + 55, y1 + 315, x2 - 55, y1 + 805)
    rounded(d, card, 34, (24, 29, 49), (126, 102, 255), 2)
    centered_text(d, (card[0], card[1] + 48, card[2], card[1] + 175), "今晚认识新朋友\n先看兴趣、语言与地区偏好", 31, (255, 255, 255), True)
    cx, cy = (card[0] + card[2]) / 2, card[1] + 285
    d.rounded_rectangle((cx - 110, cy - 54, cx + 85, cy + 54), 40, fill=(229, 252, 255))
    d.polygon([(cx + 70, cy - 35), (cx + 210, cy - 105), (cx + 210, cy + 105), (cx + 70, cy + 35)], fill=(229, 252, 255))
    d.ellipse((cx - 40, cy - 40, cx + 40, cy + 40), fill=(187, 84, 229))
    d.ellipse((cx - 16, cy - 16, cx + 16, cy + 16), fill=(7, 15, 28))
    rounded(d, (card[0] + 55, card[3] - 102, card[2] - 55, card[3] - 42), 19, (111, 83, 255))
    centered_text(d, (card[0] + 55, card[3] - 102, card[2] - 55, card[3] - 42), "邀请连线", 24, (255, 255, 255), True)
    draw_nav(d, box, 0)


def draw_permission_screen(d, box):
    x1, y1, x2, y2 = box
    d.text((x1 + 10, y1 + 2), "权限说明", font=font(34, True), fill=(255, 255, 255))
    rounded(d, (x1 + 36, y1 + 115, x2 - 36, y1 + 285), 32, (25, 33, 55))
    centered_text(d, (x1 + 60, y1 + 132, x2 - 60, y1 + 268), "双方同意后才启用\n摄像头与麦克风", 32, (255, 255, 255), True)
    for idx, (icon, title, desc, color) in enumerate([
        ("◉", "摄像头权限", "用于双方同意的一对一视讯", (92, 225, 230)),
        ("●", "麦克风权限", "用于语言交流和语音通话", (255, 94, 153)),
        ("◎", "安全提示", "可随时结束、举报或封锁", (132, 113, 255)),
    ]):
        y = y1 + 360 + idx * 150
        rounded(d, (x1 + 44, y, x2 - 44, y + 112), 28, (22, 28, 48), (53, 66, 94), 1)
        d.ellipse((x1 + 84, y + 30, x1 + 136, y + 82), fill=color)
        centered_text(d, (x1 + 84, y + 30, x1 + 136, y + 82), icon, 22, (11, 15, 29))
        d.text((x1 + 162, y + 24), title, font=font(27, True), fill=(255, 255, 255))
        d.text((x1 + 162, y + 62), desc, font=font(21), fill=(164, 176, 207))
    rounded(d, (x1 + 80, y1 + 910, x2 - 80, y1 + 976), 24, (111, 83, 255))
    centered_text(d, (x1 + 80, y1 + 910, x2 - 80, y1 + 976), "继续浏览", 26, (255, 255, 255), True)
    draw_nav(d, box, 0)


def draw_chat_screen(d, box):
    x1, y1, x2, y2 = box
    d.text((x1 + 8, y1 + 2), "聊天", font=font(34, True), fill=(255, 255, 255))
    rounded(d, (x1 + 24, y1 + 80, x2 - 24, y1 + 174), 28, (23, 30, 50))
    d.ellipse((x1 + 52, y1 + 104, x1 + 104, y1 + 156), fill=(92, 225, 230))
    d.text((x1 + 122, y1 + 96), "已确认：Mika", font=font(28, True), fill=(255, 255, 255))
    d.text((x1 + 122, y1 + 132), "连线中 · 兴趣相近", font=font(20), fill=(158, 171, 204))
    bubbles = [
        (False, "你好！你也喜欢旅行吗？"),
        (True, "喜欢，尤其是海边和城市散步。"),
        (False, "那我们很合拍。"),
        (True, "先语音聊一下也可以。"),
    ]
    y = y1 + 240
    for mine, msg in bubbles:
        w = 470 if len(msg) > 12 else 320
        h = 82
        if mine:
            boxb = (x2 - 70 - w, y, x2 - 70, y + h)
            fill = (107, 83, 255)
        else:
            boxb = (x1 + 70, y, x1 + 70 + w, y + h)
            fill = (29, 37, 59)
        rounded(d, boxb, 24, fill)
        d.text((boxb[0] + 24, boxb[1] + 24), msg, font=font(23), fill=(255, 255, 255))
        y += 118
    rounded(d, (x1 + 44, y1 + 840, x2 - 44, y1 + 932), 28, (24, 31, 50), (56, 68, 95), 1)
    d.text((x1 + 76, y1 + 871), "输入消息", font=font(24), fill=(142, 154, 183))
    pill(d, x2 - 235, y1 + 858, "发送", (92, 225, 230), (7, 15, 28), 28, 56)
    draw_nav(d, box, 1)


def draw_discover_screen(d, box):
    x1, y1, x2, y2 = box
    d.text((x1 + 8, y1 + 2), "探索", font=font(34, True), fill=(255, 255, 255))
    pill(d, x1 + 8, y1 + 70, "18+", (39, 50, 75))
    pill(d, x1 + 98, y1 + 70, "附近", (39, 50, 75))
    pill(d, x1 + 206, y1 + 70, "语音优先", (107, 83, 255))
    cards = [
        ("Alex", "音乐 · 电影 · 城市漫步", (92, 225, 230)),
        ("Mika", "咖啡 · 旅行 · 轻聊天", (255, 94, 153)),
        ("Sam", "运动 · 游戏 · 语言交换", (132, 113, 255)),
    ]
    for i, (name, meta, color) in enumerate(cards):
        y = y1 + 190 + i * 190
        rounded(d, (x1 + 26, y, x2 - 26, y + 150), 30, (23, 30, 50), (53, 66, 94), 1)
        d.ellipse((x1 + 82, y + 28, x1 + 174, y + 120), fill=color)
        centered_text(d, (x1 + 82, y + 28, x1 + 174, y + 120), name[0], 38, (8, 15, 28), True)
        d.text((x1 + 205, y + 34), name, font=font(31, True), fill=(255, 255, 255))
        d.text((x1 + 205, y + 78), meta, font=font(22), fill=(166, 178, 208))
        pill(d, x2 - 220, y + 52, "匹配", (92, 225, 230), (8, 15, 28), 26, 50)
    draw_nav(d, box, 1)


def draw_safety_screen(d, box):
    x1, y1, x2, y2 = box
    d.text((x1 + 8, y1 + 2), "安全中心", font=font(34, True), fill=(255, 255, 255))
    centered_text(d, (x1 + 60, y1 + 105, x2 - 60, y1 + 210), "安全工具随时可用", 35, (255, 255, 255), True)
    items = [
        ("举报", "异常行为可快速提交原因"),
        ("封锁", "不想再遇见的人一键封锁"),
        ("结束", "任何时候都可以离开通话"),
        ("隐私", "不显示真实电话或帐号资料"),
    ]
    for i, (title, desc) in enumerate(items):
        y = y1 + 280 + i * 145
        rounded(d, (x1 + 48, y, x2 - 48, y + 106), 28, (24, 31, 51), (55, 70, 99), 1)
        d.rounded_rectangle((x1 + 96, y + 28, x1 + 148, y + 80), 16, fill=(255, 94, 153) if i < 2 else (92, 225, 230))
        d.text((x1 + 175, y + 22), title, font=font(28, True), fill=(255, 255, 255))
        d.text((x1 + 175, y + 60), desc, font=font(21), fill=(166, 178, 208))
    rounded(d, (x1 + 92, y1 + 920, x2 - 92, y1 + 986), 24, (255, 94, 153))
    centered_text(d, (x1 + 92, y1 + 920, x2 - 92, y1 + 986), "举报或封锁", 26, (255, 255, 255), True)
    draw_nav(d, box, 3)


def screenshot(filename, title, subtitle, drawer):
    img = gradient((W, H), (9, 20, 48), (18, 143, 158))
    add_brand_bg(img)
    d = ImageDraw.Draw(img)
    header(img, title, subtitle)
    box = app_frame(img)
    drawer(d, box)
    img = img.convert("RGB")
    img.save(SCREENSHOT_OUT / filename, quality=96)


def feature_graphic():
    img = gradient((1024, 500), (11, 21, 55), (19, 154, 176))
    add_brand_bg(img, (92, 225, 230), (255, 94, 153))
    d = ImageDraw.Draw(img)
    rounded(d, (54, 64, 354, 436), 44, (12, 18, 38))
    if LOGO.exists():
        logo = Image.open(LOGO).convert("RGBA").resize((238, 238), Image.LANCZOS)
        img.alpha_composite(logo, (85, 99))
    text(d, (410, 106), "Interest Match Online", 52, fill=(255, 255, 255), bold=True)
    text(d, (412, 186), "Interest social and language exchange", 28, fill=(214, 225, 246))
    x = 412
    for label, color in [
        ("Interest profiles", (92, 225, 230)),
        ("Mutual 1v1", (255, 94, 153)),
        ("Safety tools", (132, 113, 255)),
    ]:
        x = pill(d, x, 266, label, color, (9, 15, 29), 20, 48) + 14
    img.convert("RGB").save(OUT / "feature-graphic.png", quality=96)


def main():
    SCREENSHOT_OUT.mkdir(parents=True, exist_ok=True)
    screenshot(
        "01-interest-profile-match.png",
        "兴趣社交与语言交流",
        "先看资料、兴趣与语言，再邀请 1v1 连线",
        draw_video_screen,
    )
    screenshot(
        "02-camera-mic-permission.png",
        "权限清楚才安心",
        "摄像头、麦克风只用于双方同意后的交流",
        draw_permission_screen,
    )
    screenshot(
        "03-live-chat.png",
        "确认后自然开聊",
        "语音、视讯与文字交流在同一个流程里完成",
        draw_chat_screen,
    )
    screenshot(
        "04-discover-profiles.png",
        "发现合适的新朋友",
        "用兴趣标签筛选更适合的聊天对象",
        draw_discover_screen,
    )
    screenshot(
        "05-safety-report-block.png",
        "安全工具不藏起来",
        "举报、封锁与结束通话都在需要时能找到",
        draw_safety_screen,
    )
    feature_graphic()


if __name__ == "__main__":
    main()
