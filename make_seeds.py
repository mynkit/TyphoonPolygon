import datetime
import glob
import json
import logging
import os
import re
from datetime import timezone
from typing import List, TypedDict

from bs4 import BeautifulSoup

logger = logging.getLogger(__name__)


class Meta(TypedDict):
    env: str
    risk_type: str
    risk_source: str
    version: str
    source_fetch_timestamp: str
    source_path: str
    created_at: str
    updated_at: str


class TyphoonCircleForecastSummary(TypedDict):
    title: str
    report_timestamp: str
    target_timestamp: str
    typhoon_name: str
    typhoon_name_kana: str
    typhoon_number: str
    case_id: str
    report_no: int


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
    # 基本情報
    typhoon_class: str
    typhoon_strength: str
    typhoon_size: str
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


class TyphoonCircleForecast(Meta):
    summary: TyphoonCircleForecastSummary
    details: list[TyphoonCircleForecastDetail]


def convert_time_type_a(input_time: str) -> str:
    # before(JST): 2022-11-11T14:32:00+09:00
    # after (UTC): 2022-11-11 05:32:00 UTC

    strp_format = "%Y-%m-%dT%H:%M:%S%z"
    strf_format = "%Y-%m-%d %H:%M:%S UTC"

    time_data = datetime.datetime.strptime(input_time, strp_format)
    time_string = time_data.astimezone(datetime.timezone.utc).strftime(strf_format)
    return time_string


def convert_time_type_b(input_time: str) -> str:
    # before(UTC): 20221111053214
    # after (UTC): 2022-11-11 05:32:14 UTC

    strp_format = "%Y%m%d%H%M%S"
    strf_format = "%Y-%m-%d %H:%M:%S UTC"

    time_data = datetime.datetime.strptime(input_time, strp_format)
    time_string = time_data.astimezone(datetime.timezone.utc).strftime(strf_format)
    return time_string


def get_soup(xml_path: str):
    with open(xml_path) as f:
        xml = f.read()
    soup = BeautifulSoup(xml, "xml")
    return soup


def parse_meta_data(target_file_name):
    now_utc = datetime.datetime.now(timezone.utc)
    now_utc_str = now_utc.strftime("%Y-%m-%d %H:%M:%S.%f %Z")
    timestamp = convert_time_type_b(target_file_name[:14])

    env = "local"
    risk_type = "typhoon_circle_forecast"
    risk_source = "jma"
    version = "2022-11-01"
    bucket_name = "resilire_local_jma_typhoon_circle_forecast_bucket"

    gcs_path = f"/{bucket_name}/{target_file_name}"

    meta = Meta(
        env=env,
        risk_type=risk_type,
        risk_source=risk_source,
        version=version,
        source_path=gcs_path,
        source_fetch_timestamp=timestamp,
        created_at=now_utc_str,
        updated_at=now_utc_str,
    )

    return meta


def parse_summary_data(soup):
    head = soup.find("Head")
    title = "台風解析・予報情報（５日予報）（Ｈ３０）"
    report_timestamp = convert_time_type_a(head.find("ReportDateTime").text.strip())
    target_timestamp = convert_time_type_a(head.find("TargetDateTime").text.strip())
    case_id = head.find("EventID").text.strip()
    report_no = head.find("Serial").text.strip()

    meteorological_info_parent = soup.find("MeteorologicalInfos")
    typhoone_basic_info = meteorological_info_parent.find("MeteorologicalInfo")
    typhoon_name = typhoone_basic_info.find("Name").text.strip()
    typhoon_name_kana = typhoone_basic_info.find("NameKana").text.strip()
    typhoon_number = typhoone_basic_info.find("Number").text.strip()

    summary = TyphoonCircleForecastSummary(
        title=title,
        typhoon_name=typhoon_name,
        typhoon_name_kana=typhoon_name_kana,
        typhoon_number=typhoon_number,
        report_timestamp=report_timestamp,
        target_timestamp=target_timestamp,
        case_id=case_id,
        report_no=int(report_no),
    )

    return summary


def parse_details_data(soup):
    details: List[TyphoonCircleForecastDetail] = []

    meteorological_info_parent = soup.find("MeteorologicalInfos")

    meteorological_infos = meteorological_info_parent.find_all("MeteorologicalInfo")

    for meteorological_info in meteorological_infos:
        target_timestamp = convert_time_type_a(
            meteorological_info.find("DateTime").text.strip()
        )
        target_timestamp_type = meteorological_info.find("DateTime").get("type")
        # 基本情報
        typhoon_class = ""
        typhoon_class_elem = meteorological_info.find(
            "jmx_eb:TyphoonClass", type="熱帯擾乱種類"
        )
        if typhoon_class_elem:
            typhoon_class = typhoon_class_elem.text.strip()

        typhoon_strength = ""
        typhoon_strength_elem = meteorological_info.find(
            "jmx_eb:IntensityClass", type="強さ階級"
        )
        if typhoon_strength_elem:
            typhoon_strength = typhoon_strength_elem.text.strip()

        typhoon_size = ""
        typhoon_size_elem = meteorological_info.find(
            "jmx_eb:AreaClass", type="大きさ階級"
        )
        if typhoon_size_elem:
            typhoon_size = typhoon_size_elem.text.strip()
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
        location_elem = center_part.find("Location")
        location = location_elem.text.strip() if location_elem is not None else ""
        direction = center_part.find("jmx_eb:Direction").text.strip()
        velocity = center_part.find("jmx_eb:Speed", unit="km/h").text.strip()
        central_pressure = center_part.find("jmx_eb:Pressure").text.strip()

        # 風の情報
        wind_part = meteorological_info.find("WindPart")
        max_wind_speed_near_the_center = (
            wind_part.find("jmx_eb:WindSpeed", type="最大風速", unit="m/s").text.strip()
            if wind_part is not None
            else "0"
        )
        instantaneous_max_wind_speed = (
            wind_part.find(
                "jmx_eb:WindSpeed", type="最大瞬間風速", unit="m/s"
            ).text.strip()
            if wind_part is not None
            else "0"
        )

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
            # 基本情報
            typhoon_class=typhoon_class,
            typhoon_strength=typhoon_strength,
            typhoon_size=typhoon_size,
            # 中心の情報
            latitude=float(latitude),
            longitude=float(longitude),
            location=location,
            direction=direction,
            velocity=int(velocity) if velocity != "" else 0,
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
    xml_paths = sorted(xml_paths)
    typhoon_circle_forecasts = []
    for xml_path in xml_paths:
        soup = get_soup(xml_path)

        meta = parse_meta_data(os.path.basename(xml_path))
        summary = parse_summary_data(soup)
        details = parse_details_data(soup)

        typhoon_circle_forecast_dict = meta | {
            "summary": summary,
            "details": details,
        }

        typhoon_circle_forecasts.append(
            TyphoonCircleForecast(**typhoon_circle_forecast_dict)
        )

    with open("output.jsonl", "w") as file:
        for item in typhoon_circle_forecasts:
            json.dump(item, file, ensure_ascii=False)
            file.write("\n")
