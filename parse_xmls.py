"""xmlファイルたちをscrapeしてjsonファイルに変換.

Requiments:
    pip install beautifulsoup4 lxml

Exmaples:
    python parse_xmls.py

"""

import datetime
import glob
import json
import logging
import os
import re
from typing import List, TypedDict

from bs4 import BeautifulSoup

logger = logging.getLogger(__name__)


class WarningArea(TypedDict):
    warning_area_type: str
    wind_speed: int
    circle_long_direction: str
    circle_long_radius: int
    circle_short_direction: str
    circle_short_radius: int


class TyphoonCircleForecastDetail(TypedDict):
    target_timestamp: str
    target_timestamp_type: str
    # 中心の情報
    latitude: float
    longitude: float
    location: str
    direction: str
    velocity: int
    central_pressure: int
    # 風の情報
    max_wind_speed_near_the_center: int
    instantaneous_max_wind_speed: int
    # WARNINGエリア
    warning_areas: List[WarningArea]


def convert_time_type_a(input_time: str) -> str:
    # before(JST): 2022-11-11T14:32:00+09:00
    # after (UTC): 2022-11-11 05:32:00 UTC

    strp_format = "%Y-%m-%dT%H:%M:%S%z"
    strf_format = "%Y-%m-%d %H:%M:%S UTC"

    time_data = datetime.datetime.strptime(input_time, strp_format)
    time_string = time_data.astimezone(datetime.timezone.utc).strftime(strf_format)
    return time_string


def parse_xml(xml_path: str):
    with open(xml_path) as f:
        xml = f.read()
    soup = BeautifulSoup(xml, "xml")

    details: List[TyphoonCircleForecastDetail] = []

    meteorological_info_parent = soup.find("MeteorologicalInfos")

    meteorological_infos = meteorological_info_parent.find_all("MeteorologicalInfo")

    for meteorological_info in meteorological_infos:
        target_timestamp = convert_time_type_a(
            meteorological_info.find("DateTime").text.strip()
        )
        target_timestamp_type = meteorological_info.find("DateTime").get("type")
        # 中心の情報
        center_part = meteorological_info.find("CenterPart")
        latlon_text = center_part.find(
            ["jmx_eb:Coordinate", "jmx_eb:BasePoint"], type="中心位置（度）"
        ).text.strip()
        latlon_match = re.findall(r"\A([+-][0-9\.]+)([+-][0-9\.]+)\/\Z", latlon_text)
        latitude, longitude = None, None
        if len(latlon_match) == 1:
            latitude, longitude = latlon_match[0]
        else:
            logger.warning(
                f"Could not retrieve latitude and longitude. (latlon_text: {latlon_text})"
            )
        location = center_part.find("Location").text.strip()
        direction = center_part.find("jmx_eb:Direction").text.strip()
        velocity = center_part.find("jmx_eb:Speed", unit="km/h").text.strip()
        central_pressure = center_part.find("jmx_eb:Pressure").text.strip()

        # 風の情報
        wind_part = meteorological_info.find("WindPart")
        max_wind_speed_near_the_center = wind_part.find(
            "jmx_eb:WindSpeed", type="最大風速", unit="m/s"
        ).text.strip()
        instantaneous_max_wind_speed = wind_part.find(
            "jmx_eb:WindSpeed", type="最大瞬間風速", unit="m/s"
        ).text.strip()

        # WARNINGエリア
        warning_areas: List[WarningArea] = []
        warning_area_parts = meteorological_info.find_all(
            ["WarningAreaPart", "ProbabilityCircle"]
        )
        for warning_area_part in warning_area_parts:
            warning_area_type = warning_area_part.get("type")
            if warning_area_type == "予報円":
                wind_speed = 0
            else:
                wind_speed = warning_area_part.find(
                    "jmx_eb:WindSpeed", unit="m/s"
                ).text.strip()
            axises = warning_area_part.find_all("jmx_eb:Axis")
            circle_long_direction = axises[0].find("jmx_eb:Direction").text.strip()
            circle_long_radius = axises[0].find("jmx_eb:Radius", unit="km").text.strip()
            if len(axises) == 1:
                # ひとつしかないときは円の方向に偏りがない
                circle_short_direction = axises[0].find("jmx_eb:Direction").text.strip()
                circle_short_radius = (
                    axises[0].find("jmx_eb:Radius", unit="km").text.strip()
                )
                if circle_short_direction != "":
                    raise Exception(
                        f"circle_short_direction must be empty, but is actually: {circle_short_direction}"
                    )
            elif len(axises) == 2:
                circle_short_direction = axises[1].find("jmx_eb:Direction").text.strip()
                circle_short_radius = (
                    axises[1].find("jmx_eb:Radius", unit="km").text.strip()
                )
            else:
                raise Exception("jmx_eb:Axis elems are not found.")

            if circle_long_radius == "":
                circle_long_radius = 0
            if circle_short_radius == "":
                circle_short_radius = 0
            warning_area = WarningArea(
                warning_area_type=warning_area_type,
                wind_speed=int(wind_speed),
                circle_long_direction=circle_long_direction,
                circle_long_radius=int(circle_long_radius),
                circle_short_direction=circle_short_direction,
                circle_short_radius=int(circle_short_radius),
            )
            warning_areas.append(warning_area)

        detail = TyphoonCircleForecastDetail(
            target_timestamp=target_timestamp,
            target_timestamp_type=target_timestamp_type,
            # 中心の情報
            latitude=float(latitude),
            longitude=float(longitude),
            location=location,
            direction=direction,
            velocity=int(velocity),
            central_pressure=int(central_pressure),
            # 風の情報
            max_wind_speed_near_the_center=int(max_wind_speed_near_the_center),
            instantaneous_max_wind_speed=int(instantaneous_max_wind_speed),
            warning_areas=warning_areas,
        )
        details.append(detail)

    return details


if __name__ == "__main__":
    xml_paths = glob.glob("xml/*.xml")
    for xml_path in xml_paths:
        details = parse_xml(xml_path)
        json_path = f"json/{os.path.basename(xml_path.replace('.xml', '.json'))}"

        with open(json_path, "w") as file:
            json.dump(details, file, ensure_ascii=False, indent=4)
