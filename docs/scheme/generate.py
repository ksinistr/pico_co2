from pathlib import Path
from xml.etree import ElementTree as ET


def element(parent: ET.Element, tag: str, **attributes: object) -> ET.Element:
    return ET.SubElement(
        parent,
        tag,
        {key.replace("_", "-"): str(value) for key, value in attributes.items()},
    )


def label(
    parent: ET.Element,
    x: int,
    y: int,
    text: str,
    size: int = 22,
    color: str = "#172b40",
    anchor: str = "start",
    weight: int = 400,
) -> None:
    element(
        parent,
        "text",
        x=x,
        y=y,
        font_size=size,
        fill=color,
        text_anchor=anchor,
        font_weight=weight,
    ).text = text


def box(
    parent: ET.Element,
    x: int,
    y: int,
    width: int,
    height: int,
    fill: str = "#ffffff",
    stroke: str = "#ced9e2",
    radius: int = 16,
) -> None:
    element(
        parent,
        "rect",
        x=x,
        y=y,
        width=width,
        height=height,
        rx=radius,
        fill=fill,
        stroke=stroke,
        stroke_width=2,
    )


def wire(parent: ET.Element, points: list[tuple[int, int]], color: str) -> None:
    coordinates = " ".join(f"{x},{y}" for x, y in points)
    for stroke, width in [("#ffffff", 11), (color, 5)]:
        element(
            parent,
            "polyline",
            points=coordinates,
            fill="none",
            stroke=stroke,
            stroke_width=width,
            stroke_linejoin="round",
            stroke_linecap="round",
        )


def pad(parent: ET.Element, x: int, y: int, color: str) -> None:
    element(
        parent, "circle", cx=x, cy=y, r=7, fill="#ffffff", stroke=color, stroke_width=3
    )


def junction(parent: ET.Element, x: int, y: int, color: str) -> None:
    element(parent, "circle", cx=x, cy=y, r=7, fill=color)


def pico(parent: ET.Element, colors: dict[str, str]) -> dict[str, tuple[int, int]]:
    box(parent, 1010, 320, 290, 620, "#eaf4ef", "#638f79")
    box(parent, 1105, 291, 100, 64, "#e4e9ee", "#8596a6", 8)
    label(parent, 1155, 282, "USB · power and firmware", 20, anchor="middle")
    box(parent, 1130, 575, 50, 100, "#2c4050", "#2c4050", 6)
    label(parent, 1155, 632, "Pico", 20, "#ffffff", "middle", 700)
    box(parent, 1125, 379, 60, 28, "#ffffff", "#9aab9f", 4)
    label(parent, 1155, 432, "BOOTSEL", 13, anchor="middle")
    left = [
        "GP0",
        "GP1",
        "GND",
        "GP2",
        "GP3",
        "GP4",
        "GP5",
        "GND",
        "GP6",
        "GP7",
        "GP8",
        "GP9",
        "GND",
        "GP10",
        "GP11",
        "GP12",
        "GP13",
        "GND",
        "GP14",
        "GP15",
    ]
    right = [
        "VBUS",
        "VSYS",
        "GND",
        "3V3_EN",
        "3V3 OUT",
        "ADC_VREF",
        "GP28",
        "AGND",
        "GP27",
        "GP26",
        "RUN",
        "GP22",
        "GND",
        "GP21",
        "GP20",
        "GP19",
        "GP18",
        "GND",
        "GP17",
        "GP16",
    ]
    selected = {6: "SDA", 7: "SCL", 14: "LEFT", 15: "RIGHT", 36: "3V3", 38: "GND"}
    result = {}
    for names, x, start, step, anchor, offset in [
        (left, 1010, 1, 1, "start", 18),
        (right, 1300, 40, -1, "end", -18),
    ]:
        for index, name in enumerate(names):
            number = start + index * step
            y = 365 + index * 27
            net = selected.get(number)
            color = colors[net] if net else "#93a69d"
            pad(parent, x, y, color)
            label(
                parent,
                x + offset,
                y + 5,
                f"{number:02}  {name}",
                16,
                color,
                anchor,
                700 if net else 400,
            )
            if net:
                result[net] = (x, y)
    label(
        parent, 1155, 976, "Top view · USB at the top", 22, anchor="middle", weight=700
    )
    label(parent, 1155, 1007, "Numbers 1–40 are physical pins", 19, anchor="middle")
    return result


def module(
    parent: ET.Element,
    y: int,
    title: str,
    subtitle: str,
    kind: str,
    colors: dict[str, str],
) -> list[tuple[str, int, int]]:
    button = kind == "button"
    x = 945 if button else 60
    height = 150 if button else 230 if kind == "display" else 215
    box(parent, x, y, 445 if button else 450, height, "#ffffff")
    title_y, subtitle_y = (188, 215) if kind == "display" else (37, 65)
    label(parent, x + 24, y + title_y, title, 27, weight=700)
    label(parent, x + 24, y + subtitle_y, subtitle, 19, "#526679")
    if button:
        box(parent, x + 310, y + 83, 72, 48, "#2c4050", "#2c4050", 5)
        label(parent, x + 346, y + 112, "TTP223", 15, "#ffffff", "middle")
        ports = [
            ("GND", "GND", x, y + 85),
            (title.split()[0], "I/O", x, y + 112),
            ("3V3", "VCC", x, y + 139),
        ]
    elif kind == "display":
        ports = [
            (net, name, 140 + index * 100, y + 30)
            for index, (net, name) in enumerate(
                [("GND", "GND"), ("3V3", "VCC"), ("SCL", "SCL"), ("SDA", "SDA")]
            )
        ]
        box(parent, 84, y + 80, 402, 75, "#142b3d", "#142b3d", 8)
        label(parent, 104, y + 111, "12:45                 CO₂ 800 ppm", 21, "#d7f1ff")
        label(parent, 104, y + 140, "24°C   45%", 18, "#d7f1ff")
    elif kind == "rtc":
        ports = [
            ("SCL", "SCL", 510, y + 126),
            ("SDA", "SDA", 510, y + 149),
            ("3V3", "VCC", 510, y + 172),
            ("GND", "GND", 510, y + 195),
        ]
        element(
            parent,
            "circle",
            cx=245,
            cy=y + 143,
            r=48,
            fill="#edf1f4",
            stroke="#b8c4ce",
            stroke_width=2,
        )
        label(parent, 245, y + 140, "CR2032", 19, "#617385", "middle", 700)
        label(parent, 245, y + 163, "+3 V", 18, "#617385", "middle")
        label(parent, 84, y + 101, "Spare header", 16, "#71808d")
        for _, name, _, port_y in ports:
            pad(parent, 60, port_y, "#71808d")
            label(parent, 80, port_y + 7, name, 20, "#71808d")
        for index, name in enumerate(["32K", "SQW"]):
            port_y = y + 80 + index * 23
            label(parent, 490, port_y + 7, name, 21, "#71808d", "end", 700)
            pad(parent, 510, port_y, "#71808d")
            element(
                parent,
                "path",
                d=f"M 523,{port_y - 5} l 10,10 m -10,0 l 10,-10",
                fill="none",
                stroke="#71808d",
                stroke_width=2,
            )
            label(parent, 542, port_y + 6, "NC", 16, "#71808d")
    elif kind == "scd40":
        ports = [
            (net, name, 140 + index * 100, y + 190)
            for index, (net, name) in enumerate(
                [("GND", "GND"), ("3V3", "VDD"), ("SCL", "SCL"), ("SDA", "SDA")]
            )
        ]
        box(parent, 218, y + 82, 134, 78, "#d7dadd", "#9ba1a6", 12)
        box(parent, 247, y + 99, 76, 44, "#929b9e", "#929b9e", 5)
    else:
        ports = [
            ("3V3", "3V3", 150, y + 190),
            ("GND", "GND", 200, y + 190),
            ("SCL", "SCL", 250, y + 190),
            ("SDA", "SDA", 300, y + 190),
            ("3V3", "ADD", 350, y + 190),
            ("3V3", "CS", 400, y + 190),
        ]
        for port_x, name in [(100, "VIN"), (450, "NT/INT")]:
            port_y = y + 190
            label(parent, port_x, port_y - 17, name, 16, "#71808d", "middle", 700)
            label(parent, port_x, port_y - 44, "NC", 16, "#71808d", "middle")
            pad(parent, port_x, port_y, "#71808d")
            element(
                parent,
                "path",
                d=f"M {port_x - 5},{port_y + 10} l 10,10 m -10,0 l 10,-10",
                fill="none",
                stroke="#71808d",
                stroke_width=2,
            )
        for offset, name in [(0, "ENS160"), (104, "AHT20")]:
            box(parent, 170 + offset, y + 94, 92, 55, "#eae5f4", "#b5a5ca", 7)
            label(parent, 216 + offset, y + 128, name, 18, "#594578", "middle", 700)
    for net, name, port_x, port_y in ports:
        if kind == "display":
            label(parent, port_x, port_y + 30, name, 21, colors[net], "middle", 700)
        elif kind in ["scd40", "combined"]:
            label(
                parent,
                port_x,
                port_y - 17,
                name,
                16 if kind == "combined" else 21,
                colors[net],
                "middle",
                700,
            )
        elif button:
            label(parent, port_x + 20, port_y + 7, name, 21, colors[net], weight=700)
        else:
            label(parent, port_x - 20, port_y + 7, name, 21, colors[net], "end", 700)
        pad(parent, port_x, port_y, colors[net])
    return [(net, port_x, port_y) for net, _, port_x, port_y in ports]


def render(sensor: str, colors: dict[str, str], rails: dict[str, int]) -> str:
    combined = sensor == "aht20-ens160"
    title = "AHT20 + ENS160" if combined else "SCD40"
    svg = ET.Element(
        "svg",
        {
            "xmlns": "http://www.w3.org/2000/svg",
            "viewBox": "0 0 1440 1720",
            "width": "1440",
            "height": "1720",
            "style": "width:100%;height:auto;background:#f5f8fb",
            "font-family": "DejaVu Sans, Arial, sans-serif",
            "role": "img",
            "aria-labelledby": "title description",
        },
    )
    element(
        svg, "title", id="title"
    ).text = f"Raspberry Pi Pico + {title} wiring diagram"
    element(svg, "desc", id="description").text = (
        "3V3 OUT pin 36 to all power inputs (VCC/3V3/VDD); GND pin 38 to all GND; "
        "GP4 pin 6 to all SDA; GP5 pin 7 to all SCL; "
        "GP10 pin 14 to left button OUT; GP11 pin 15 to right button OUT. "
        "Includes SSD1306, DS3231, sensors and two touch buttons."
    )
    if combined:
        svg.find(
            "desc"
        ).text += " ENS160 ADD and CS to 3.3 V; VIN and NT/INT unconnected."
    box(svg, 0, 0, 1440, 1720, "#f5f8fb", "#f5f8fb", 0)
    label(svg, 60, 70, f"Pico + {title}", 44, weight=700)
    label(
        svg,
        60,
        111,
        "Soldering guide · shared I²C bus · 3.3 V module power",
        25,
        "#526679",
    )
    for net, x in rails.items():
        wire(svg, [(x, 242), (x, 1450)], colors[net])
    pins = pico(svg, colors)
    wire(
        svg,
        [
            pins["3V3"],
            (1360, pins["3V3"][1]),
            (1360, 160),
            (rails["3V3"], 160),
            (rails["3V3"], 242),
        ],
        colors["3V3"],
    )
    wire(
        svg,
        [
            pins["GND"],
            (1390, pins["GND"][1]),
            (1390, 195),
            (rails["GND"], 195),
            (rails["GND"], 242),
        ],
        colors["GND"],
    )
    for net in ["SDA", "SCL"]:
        x, y = pins[net]
        wire(svg, [(x, y), (rails[net], y)], colors[net])
        junction(svg, rails[net], y, colors[net])
        label(
            svg,
            865,
            y - 12,
            f"{net} · GP{4 if net == 'SDA' else 5}",
            20,
            colors[net],
            "middle",
            700,
        )
    sensor_subtitle = (
        "Component side | 0x38 + 0x53" if combined else "Sensor side | I²C 0x62"
    )
    modules = [
        (320, "OLED SSD1306 · 128 × 32", "Front view | I²C 0x3C", "display"),
        (580, "DS3231 clock", "Battery side | I²C 0x68", "rtc"),
        (835, title, sensor_subtitle, "combined" if combined else "scd40"),
        (1130, "LEFT · TTP223", "Component side · active HIGH", "button"),
        (1300, "RIGHT · TTP223", "Component side · active HIGH", "button"),
    ]
    for y, name, subtitle, kind in modules:
        ports = module(svg, y, name, subtitle, kind, colors)
        for index, (net, port_x, port_y) in enumerate(ports):
            if net in rails:
                if kind in ["display", "scd40", "combined"]:
                    if kind == "display":
                        track_y = 250 + index * 18
                    elif kind == "combined":
                        track_y = 1115 - index * 12
                    else:
                        track_y = 1110 - index * 18
                    points = [
                        (port_x, port_y),
                        (port_x, track_y),
                        (rails[net], track_y),
                    ]
                else:
                    track_y = port_y
                    points = [(port_x, port_y), (rails[net], port_y)]
                wire(svg, points, colors[net])
                junction(svg, rails[net], track_y, colors[net])
            else:
                route_x = 860 if net == "LEFT" else 900
                wire(
                    svg,
                    [
                        (port_x, port_y),
                        (route_x, port_y),
                        (route_x, pins[net][1]),
                        pins[net],
                    ],
                    colors[net],
                )
    for net, x in rails.items():
        box(svg, x - 24, 210, 48, 30, "#f5f8fb", "#f5f8fb", 0)
        label(svg, x, 231, net, 19, colors[net], "middle", 700)
    box(svg, 60, 1130, 450, 320, "#ffffff")
    label(svg, 84, 1170, "Pico soldering checklist", 25, weight=700)
    for index, (net, name) in enumerate(
        [
            ("3V3", "36 · 3V3 OUT → 3V3/VCC" if combined else "36 · 3V3 OUT → VCC/VDD"),
            ("GND", "38 · GND → all GND"),
            ("SDA", "6 · GP4 → all SDA"),
            ("SCL", "7 · GP5 → all SCL"),
            ("LEFT", "14 · GP10 → left I/O"),
            ("RIGHT", "15 · GP11 → right I/O"),
        ]
    ):
        label(svg, 84, 1212 + index * 38, name, 22, colors[net], weight=700)
    junction(svg, 71, 1492, "#172b40")
    label(svg, 88, 1499, "Dot = solder together.", 21, weight=700)
    wire(svg, [(415, 1492), (475, 1492)], colors["SDA"])
    wire(svg, [(445, 1477), (445, 1507)], colors["SCL"])
    label(
        svg,
        491,
        1499,
        "Crossing without a dot = no connection. NC = leave unconnected.",
        21,
    )
    box(svg, 60, 1530, 1330, 155, "#eaf0f6", "#eaf0f6")
    label(
        svg,
        84,
        1564,
        "Match the indicated board views and the labels on your own modules.",
        23,
        weight=700,
    )
    label(
        svg,
        84,
        1597,
        "Use breakout modules rated for a 3.3 V power input, with I²C pull-ups to 3.3 V.",
        22,
    )
    label(
        svg,
        84,
        1630,
        "Buttons: momentary mode, OUT = HIGH on touch. Leave unused Pico pins disconnected.",
        22,
    )
    note = (
        "ENS160: ADD + CS → 3V3; VIN + NT/INT → NC. I²C address 0x53."
        if combined
        else "SCD40: up to 205 mA peak at 3.3 V. USB supplies power through the Pico."
    )
    label(svg, 84, 1663, note, 22)
    ET.indent(svg)
    return ET.tostring(svg, encoding="unicode", xml_declaration=True) + "\n"


def main() -> None:
    colors = {
        "3V3": "#c83637",
        "GND": "#344554",
        "SDA": "#126db1",
        "SCL": "#b47108",
        "LEFT": "#8051ac",
        "RIGHT": "#007e76",
    }
    rails = {"3V3": 610, "GND": 660, "SDA": 710, "SCL": 760}
    output = Path(__file__).resolve().parent
    for sensor in ["scd40", "aht20-ens160"]:
        path = output / f"pico-{sensor}.svg"
        path.write_text(render(sensor, colors, rails), encoding="utf-8")
        print(path)


if __name__ == "__main__":
    main()
