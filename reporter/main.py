import base64
import os
import sys

# Add the proto directory to the Python path
sys.path.append(os.path.abspath(os.path.join(os.path.dirname(__file__), '..', 'proto', 'gen', 'python')))

from google.cloud import firestore
from nhd_pb2 import ReportRun

PROJECT_ID = os.environ.get("GOOGLE_CLOUD_PROJECT")
db = firestore.Client()

def handle_report_request(event, context):
    """Triggered from a message on a Cloud Pub/Sub topic.
    Args:
         event (dict): Event payload.
         context (google.cloud.functions.Context): Metadata for the event.
    """
    report_run_id = base64.b64decode(event['data']).decode('utf-8')
    print(f"Processing report run: {report_run_id}")

    report_run_ref = db.collection('report_runs').document(report_run_id)
    report_run_ref.update({
        'status': ReportRun.COMPLETED
    })

    print(f"Report run {report_run_id} completed.")

