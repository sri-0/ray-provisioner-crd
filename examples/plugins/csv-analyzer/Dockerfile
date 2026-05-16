FROM registry.example.com/ray-framework:latest

WORKDIR /opt/ray
COPY requirements.txt /tmp/plugin-requirements.txt
RUN pip install --no-cache-dir -r /tmp/plugin-requirements.txt

COPY plugin.py /opt/ray/plugins/csv_analyzer/plugin.py
ENV PYTHONPATH=/opt/ray
