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

# Load hazard zones from mock data
fire_hazard_zones = geopandas.read_file(os.path.join(os.path.dirname(__file__), 'data', 'mock_fire_hazard_zones.geojson'))

def handle_report_request(event, context):
    """Triggered from a message on a Cloud Pub/Sub topic.
    Args:
         event (dict): Event payload.
         context (google.cloud.functions.Context): Metadata for the event.
    """
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
        return

    property_address_ref = db.collection('property_addresses').document(property_address_id)
    property_address_doc = property_address_ref.get()

    if not property_address_doc.exists:
        print(f"Error: PropertyAddress document {property_address_id} not found.")
        return

    property_address_data = property_address_doc.to_dict()
    coordinates = property_address_data.get('coordinates')

    if not coordinates or 'latitude' not in coordinates or 'longitude' not in coordinates:
        print(f"Error: Coordinates not found in PropertyAddress {property_address_id}")
        return

    # Perform Point-in-Polygon (PIP) analysis
    property_location = Point(coordinates['longitude'], coordinates['latitude'])
    
    # Initialize results
    hazard_results = {
        'in_very_high_fire_hazard_severity_zone': False,
        # Initialize other hazards to False
        'in_special_flood_hazard_area': False,
        'in_dam_inundation_area': False,
        'in_wildland_fire_area': False,
        'in_earthquake_fault_zone': False,
        'in_seismic_hazard_zone': False,
    }

    # Check against fire hazard zones
    if any(fire_hazard_zones.geometry.contains(property_location)):
        hazard_results['in_very_high_fire_hazard_severity_zone'] = True

    # Update the Firestore document
    report_run_ref.update({
        'status': ReportRun.COMPLETED,
        'results': hazard_results
    })

    print(f"Report run {report_run_id} completed with results: {hazard_results}")

