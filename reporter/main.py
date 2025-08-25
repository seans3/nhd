import base64
import os
import sys
import geopandas
from shapely.geometry import Point

# Add the proto directory to the Python path
sys.path.append(os.path.abspath(os.path.join(os.path.dirname(__file__), '..', 'proto', 'gen', 'python')))

from google.cloud import firestore
from nhd_pb2 import ReportRun

PROJECT_ID = os.environ.get("GOOGLE_CLOUD_PROJECT")
db = firestore.Client()

# ====================================================================================
# LOAD HAZARD ZONE DATA
# In a production environment, this data would be read from a more robust and
# optimized source like PostGIS, BigQuery, or spatially-indexed flat files (e.g., Parquet).
# For this implementation, we load the mock GeoJSON files on startup.
# ====================================================================================
def load_zone_data(filename):
    """Loads a GeoJSON file from the data directory."""
    path = os.path.join(os.path.dirname(__file__), 'data', filename)
    return geopandas.read_file(path)

try:
    flood_zones = load_zone_data('mock_flood_hazard_zones.geojson')
    dam_zones = load_zone_data('mock_dam_inundation_areas.geojson')
    fire_zones = load_zone_data('mock_fire_hazard_zones.geojson')
    wildland_zones = load_zone_data('mock_wildland_fire_areas.geojson')
    fault_zones = load_zone_data('mock_earthquake_fault_zones.geojson')
    seismic_zones = load_zone_data('mock_seismic_hazard_zones.geojson')
    print("All mock hazard zone data loaded successfully.")
except Exception as e:
    print(f"FATAL: Could not load mock hazard zone data. Error: {e}")
    # In a real Cloud Function, you might want to handle this more gracefully.
    # For this example, we'll let it fail on startup if data is missing.

def handle_report_request(event, context):
    """Triggered from a message on a Cloud Pub/Sub topic."""
    report_run_id = base64.b64decode(event['data']).decode('utf-8')
    print(f"Processing report run: {report_run_id}")

    report_run_ref = db.collection('report_runs').document(report_run_id)
    report_run_doc = report_run_ref.get()

    if not report_run_doc.exists:
        print(f"Error: ReportRun document {report_run_id} not found.")
        return

    report_run_data = report_run_doc.to_dict()
    property_address_id = report_run_data.get('property_address_id')

    if not property_address_id:
        print(f"Error: property_address_id not found in ReportRun {report_run_id}")
        report_run_ref.update({'status': ReportRun.FAILED, 'results': {'error': 'Missing property_address_id'}})
        return

    property_address_ref = db.collection('property_addresses').document(property_address_id)
    property_address_doc = property_address_ref.get()

    if not property_address_doc.exists:
        print(f"Error: PropertyAddress document {property_address_id} not found.")
        report_run_ref.update({'status': ReportRun.FAILED, 'results': {'error': 'PropertyAddress not found'}})
        return

    property_address_data = property_address_doc.to_dict()
    coordinates = property_address_data.get('coordinates')

    if not coordinates or 'latitude' not in coordinates or 'longitude' not in coordinates:
        print(f"Error: Coordinates not found in PropertyAddress {property_address_id}")
        report_run_ref.update({'status': ReportRun.FAILED, 'results': {'error': 'Coordinates not found'}})
        return

    # --- Perform Full Point-in-Polygon (PIP) Analysis ---
    property_location = Point(coordinates['longitude'], coordinates['latitude'])
    
    hazard_results = {
        'in_special_flood_hazard_area': any(flood_zones.geometry.contains(property_location)),
        'in_dam_inundation_area': any(dam_zones.geometry.contains(property_location)),
        'in_very_high_fire_hazard_severity_zone': any(fire_zones.geometry.contains(property_location)),
        'in_wildland_fire_area': any(wildland_zones.geometry.contains(property_location)),
        'in_earthquake_fault_zone': any(fault_zones.geometry.contains(property_location)),
        'in_seismic_hazard_zone': any(seismic_zones.geometry.contains(property_location)),
    }

    # Update the Firestore document with the complete results
    report_run_ref.update({
        'status': ReportRun.COMPLETED,
        'results': hazard_results
    })

    print(f"Report run {report_run_id} completed with results: {hazard_results}")

