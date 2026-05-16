FROM rayproject/ray:2.31.0-py311

WORKDIR /opt/ray/framework
COPY pyproject.toml ./
COPY ray_framework ./ray_framework
RUN pip install --no-cache-dir .

ENV PYTHONPATH=/opt/ray/framework
