#!/usr/bin/env python3
"""
Phase 8: Interactive ECharts Dashboard (self-contained HTML)
"""

import json
import sys
from collections import Counter, defaultdict

NAVY = "#1F4E79"
GREEN = "#27AE60"
LIGHT_GREEN = "#A9D18E"
YELLOW = "#F1C40F"
GRAY = "#BDC3C7"
RED = "#E74C3C"
ORANGE = "#F39C12"


def main():
    with open("/tmp/nyc_assemblage/verified_properties.json") as f:
        properties = json.load(f)

    with open("/tmp/nyc_assemblage/clusters.json") as f:
        clusters = json.load(f)

    # Build chart data
    # 1. Scatter map data (lat/lng/score/address/price)
    scatter_data = []
    for p in properties:
        lat = p.get("latitude")
        lng = p.get("longitude")
        score = p.get("_score", 0)
        addr = p.get("address", "N/A")
        price = p.get("price", 0) or 0
        borough = p.get("_borough", "N/A")
        priority = p.get("_priority", "Watchlist")
        zoning = p.get("_zoning", "N/A")
        hpd = p.get("_hpd_violations", 0) or 0
        tax_lien = "Yes" if p.get("_tax_lien") else "No"
        lis = "Yes" if p.get("_acris_lis_pendens") else "No"
        if lat and lng:
            scatter_data.append({
                "value": [round(float(lng), 5), round(float(lat), 5), score],
                "name": addr,
                "price": price,
                "borough": borough,
                "priority": priority,
                "zoning": zoning,
                "hpd": hpd,
                "tax_lien": tax_lien,
                "lis_pendens": lis,
            })

    # 2. Score distribution
    score_bins = {"0-4": 0, "5-9": 0, "10-14": 0, "15-19": 0, "20+": 0}
    for p in properties:
        s = p.get("_score", 0)
        if s >= 20:
            score_bins["20+"] += 1
        elif s >= 15:
            score_bins["15-19"] += 1
        elif s >= 10:
            score_bins["10-14"] += 1
        elif s >= 5:
            score_bins["5-9"] += 1
        else:
            score_bins["0-4"] += 1

    # 3. Signal frequency
    signal_freq = {
        "Tax Liens": sum(1 for p in properties if p.get("_tax_lien")),
        "HPD 5+ Violations": sum(1 for p in properties if (p.get("_hpd_violations") or 0) >= 5),
        "Lis Pendens": sum(1 for p in properties if p.get("_acris_lis_pendens")),
        "Federal Lien": sum(1 for p in properties if p.get("_acris_federal_lien")),
        "Estate/Probate": sum(1 for p in properties if p.get("_acris_estate")),
        "DOM > 180": sum(1 for p in properties if (p.get("daysOnMarket") or 0) > 180),
        "Demo on Block": sum(1 for p in properties if p.get("_block_demo")),
        "LLC Deed on Block": sum(1 for p in properties if p.get("_block_llc_deed")),
        "In Cluster": sum(1 for p in properties if p.get("_in_cluster")),
        "SE Price Drop >10%": sum(1 for p in properties if (p.get("se_price_drop_pct") or 0) > 10),
    }
    signal_freq = {k: v for k, v in signal_freq.items() if v > 0}

    # 4. Borough breakdown by priority
    borough_priority = defaultdict(lambda: defaultdict(int))
    for p in properties:
        b = p.get("_borough", "Unknown")
        priority = p.get("_priority", "Watchlist")
        borough_priority[b][priority] += 1

    boroughs = sorted(borough_priority.keys())
    priorities = ["Immediate", "High", "Moderate", "Watchlist"]

    # 5. Price vs Score scatter (for all properties with price)
    price_score_data = defaultdict(list)
    for p in properties:
        price = p.get("price", 0) or 0
        score = p.get("_score", 0)
        borough = p.get("_borough", "Unknown")
        addr = p.get("address", "")
        if price > 0 and score > 0:
            price_score_data[borough].append({
                "value": [price, score],
                "name": addr
            })

    # 6. Year built histogram
    year_bins = defaultdict(int)
    for p in properties:
        try:
            yr = int(float(p.get("_year_built", 0) or 0))
            if 1800 < yr <= 2026:
                decade = (yr // 10) * 10
                year_bins[decade] += 1
        except (ValueError, TypeError):
            pass

    year_data = sorted(year_bins.items())

    # 7. Zoning distribution
    zone_counts = Counter(p.get("_zoning", "Unknown") for p in properties)
    top_zones = zone_counts.most_common(12)

    # Color map for boroughs
    borough_colors = {
        "Bronx": "#E74C3C",
        "Brooklyn": "#3498DB",
        "Manhattan": "#9B59B6",
        "Queens": "#1ABC9C",
    }

    # Serialize data for JS
    scatter_json = json.dumps(scatter_data)
    price_score_json = json.dumps(dict(price_score_data))

    html = f"""<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <title>NYC Assemblage Intelligence Dashboard — 2026-03-24</title>
  <script src="https://cdn.jsdelivr.net/npm/echarts@5/dist/echarts.min.js"></script>
  <style>
    * {{ box-sizing: border-box; margin: 0; padding: 0; }}
    body {{
      font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Arial, sans-serif;
      background: #0F1B2D;
      color: #E8EDF3;
      padding: 24px;
    }}
    h1 {{
      color: #FFFFFF;
      font-size: 26px;
      margin-bottom: 6px;
      font-weight: 700;
    }}
    .subtitle {{
      color: #7FB3D3;
      font-size: 14px;
      margin-bottom: 24px;
    }}
    .stats-row {{
      display: flex;
      gap: 16px;
      margin-bottom: 24px;
      flex-wrap: wrap;
    }}
    .stat-card {{
      background: #1A2B45;
      border: 1px solid #2E4A6E;
      border-radius: 8px;
      padding: 16px 24px;
      flex: 1;
      min-width: 140px;
      text-align: center;
    }}
    .stat-card .value {{
      font-size: 32px;
      font-weight: 700;
      color: #4FC3F7;
    }}
    .stat-card.green .value {{ color: #27AE60; }}
    .stat-card.yellow .value {{ color: #F1C40F; }}
    .stat-card.red .value {{ color: #E74C3C; }}
    .stat-card .label {{
      font-size: 12px;
      color: #7FB3D3;
      margin-top: 4px;
      text-transform: uppercase;
      letter-spacing: 0.5px;
    }}
    .chart-container {{
      background: #1A2B45;
      border: 1px solid #2E4A6E;
      border-radius: 10px;
      margin-bottom: 20px;
      padding: 20px;
    }}
    .chart-container h2 {{
      color: #4FC3F7;
      font-size: 16px;
      margin-bottom: 16px;
      font-weight: 600;
    }}
    .chart {{ width: 100%; height: 450px; }}
    .chart-tall {{ width: 100%; height: 600px; }}
    .grid-2 {{
      display: grid;
      grid-template-columns: 1fr 1fr;
      gap: 20px;
    }}
    .grid-3 {{
      display: grid;
      grid-template-columns: 1fr 1fr 1fr;
      gap: 20px;
    }}
    footer {{
      margin-top: 32px;
      text-align: center;
      color: #4A6E8A;
      font-size: 12px;
    }}
  </style>
</head>
<body>
  <h1>NYC Assemblage Intelligence Dashboard</h1>
  <div class="subtitle">Generated 2026-03-24 | R7+ Residential Development Sites | Pre-Market Distress Signal Analysis</div>

  <!-- Stats Cards -->
  <div class="stats-row">
    <div class="stat-card green">
      <div class="value">{sum(1 for p in properties if p.get('_priority') == 'Immediate')}</div>
      <div class="label">Immediate Priority</div>
    </div>
    <div class="stat-card">
      <div class="value">{sum(1 for p in properties if p.get('_priority') == 'High')}</div>
      <div class="label">High Priority</div>
    </div>
    <div class="stat-card yellow">
      <div class="value">{sum(1 for p in properties if p.get('_priority') == 'Moderate')}</div>
      <div class="label">Moderate Priority</div>
    </div>
    <div class="stat-card">
      <div class="value">{len(properties)}</div>
      <div class="label">R7+ Properties</div>
    </div>
    <div class="stat-card red">
      <div class="value">{sum(1 for p in properties if p.get('_tax_lien'))}</div>
      <div class="label">Tax Liens</div>
    </div>
    <div class="stat-card red">
      <div class="value">{sum(1 for p in properties if p.get('_acris_lis_pendens'))}</div>
      <div class="label">Lis Pendens</div>
    </div>
    <div class="stat-card">
      <div class="value">{len(clusters)}</div>
      <div class="label">Block Clusters</div>
    </div>
    <div class="stat-card">
      <div class="value">{sum(1 for p in properties if p.get('_acris_estate'))}</div>
      <div class="label">Estate Signals</div>
    </div>
  </div>

  <!-- Geo Map -->
  <div class="chart-container">
    <h2>Property Locations — Sized & Colored by Composite Score</h2>
    <div id="geoChart" class="chart-tall"></div>
  </div>

  <!-- Grid: Score Dist + Signal Freq -->
  <div class="grid-2">
    <div class="chart-container">
      <h2>Score Distribution</h2>
      <div id="scoreChart" class="chart"></div>
    </div>
    <div class="chart-container">
      <h2>Distress Signals Detected</h2>
      <div id="signalChart" class="chart"></div>
    </div>
  </div>

  <!-- Borough Breakdown -->
  <div class="chart-container">
    <h2>R7+ Properties by Borough and Priority Tier</h2>
    <div id="boroughChart" class="chart"></div>
  </div>

  <!-- Price vs Score -->
  <div class="chart-container">
    <h2>Asking Price vs. Composite Score (High-Score / Low-Price Outliers)</h2>
    <div id="priceScoreChart" class="chart"></div>
  </div>

  <!-- Grid: Year Built + Zoning -->
  <div class="grid-2">
    <div class="chart-container">
      <h2>Year Built Distribution</h2>
      <div id="yearChart" class="chart"></div>
    </div>
    <div class="chart-container">
      <h2>Zoning District Distribution</h2>
      <div id="zoneChart" class="chart"></div>
    </div>
  </div>

  <footer>
    NYC Assemblage Intelligence Report — 2026-03-24 | Data sources: Zillow, NYC PLUTO, ACRIS, DOB, HPD, NYC Finance, StreetEasy
  </footer>

  <script>
    var scatterRaw = {scatter_json};
    var priceScoreRaw = {price_score_json};

    // ---- GEO SCATTER ----
    var geoChart = echarts.init(document.getElementById('geoChart'));
    var priorityColorMap = {{
      'Immediate': '#27AE60',
      'High': '#A9D18E',
      'Moderate': '#F1C40F',
      'Watchlist': '#7FB3D3'
    }};
    var geoData = scatterRaw.map(function(p) {{
      return {{
        value: p.value,
        name: p.name,
        itemStyle: {{ color: priorityColorMap[p.priority] || '#7FB3D3' }},
        tooltip_extra: p
      }};
    }});
    geoChart.setOption({{
      backgroundColor: '#1A2B45',
      title: {{ text: '', left: 'center' }},
      tooltip: {{
        trigger: 'item',
        backgroundColor: '#0F1B2D',
        borderColor: '#2E4A6E',
        textStyle: {{ color: '#E8EDF3', fontSize: 12 }},
        formatter: function(params) {{
          var d = params.data.tooltip_extra || {{}};
          return '<strong>' + params.name + '</strong><br/>' +
            'Score: <b>' + params.value[2] + '</b> (' + (d.priority||'') + ')<br/>' +
            'Price: $' + ((d.price||0).toLocaleString()) + '<br/>' +
            'Zoning: ' + (d.zoning||'N/A') + '<br/>' +
            'Borough: ' + (d.borough||'N/A') + '<br/>' +
            'HPD Violations: ' + (d.hpd||0) + '<br/>' +
            'Tax Lien: ' + (d.tax_lien||'No') + '<br/>' +
            'Lis Pendens: ' + (d.lis_pendens||'No');
        }}
      }},
      legend: {{
        data: ['Immediate', 'High', 'Moderate', 'Watchlist'],
        textStyle: {{ color: '#E8EDF3' }},
        top: 10
      }},
      xAxis: {{
        name: 'Longitude', min: -74.05, max: -73.70,
        nameTextStyle: {{ color: '#7FB3D3' }},
        axisLabel: {{ color: '#7FB3D3' }},
        splitLine: {{ lineStyle: {{ color: '#2E4A6E' }} }}
      }},
      yAxis: {{
        name: 'Latitude', min: 40.55, max: 40.95,
        nameTextStyle: {{ color: '#7FB3D3' }},
        axisLabel: {{ color: '#7FB3D3' }},
        splitLine: {{ lineStyle: {{ color: '#2E4A6E' }} }}
      }},
      series: ['Immediate', 'High', 'Moderate', 'Watchlist'].map(function(tier) {{
        return {{
          name: tier,
          type: 'scatter',
          data: geoData.filter(function(d) {{ return d.tooltip_extra && d.tooltip_extra.priority === tier; }}),
          symbolSize: function(val) {{ return Math.max(val[2] * 2.5, 6); }},
          itemStyle: {{ color: priorityColorMap[tier], opacity: 0.85 }}
        }};
      }})
    }});

    // ---- SCORE DISTRIBUTION ----
    var scoreChart = echarts.init(document.getElementById('scoreChart'));
    scoreChart.setOption({{
      backgroundColor: '#1A2B45',
      tooltip: {{ trigger: 'axis', backgroundColor: '#0F1B2D', borderColor: '#2E4A6E', textStyle: {{ color: '#E8EDF3' }} }},
      xAxis: {{
        type: 'category',
        data: {json.dumps(list(score_bins.keys()))},
        axisLabel: {{ color: '#7FB3D3' }},
        axisLine: {{ lineStyle: {{ color: '#2E4A6E' }} }}
      }},
      yAxis: {{
        type: 'value',
        axisLabel: {{ color: '#7FB3D3' }},
        splitLine: {{ lineStyle: {{ color: '#2E4A6E' }} }}
      }},
      series: [{{
        type: 'bar',
        data: {json.dumps(list(score_bins.values()))},
        itemStyle: {{
          color: function(params) {{
            var colors = ['#7FB3D3', '#A9D18E', '#F1C40F', '#F39C12', '#27AE60'];
            return colors[params.dataIndex];
          }}
        }},
        label: {{ show: true, position: 'top', color: '#E8EDF3' }}
      }}]
    }});

    // ---- SIGNAL FREQUENCY ----
    var signalChart = echarts.init(document.getElementById('signalChart'));
    signalChart.setOption({{
      backgroundColor: '#1A2B45',
      tooltip: {{ trigger: 'item', backgroundColor: '#0F1B2D', borderColor: '#2E4A6E', textStyle: {{ color: '#E8EDF3' }} }},
      legend: {{ orient: 'vertical', right: 10, top: 'center', textStyle: {{ color: '#E8EDF3', fontSize: 11 }} }},
      series: [{{
        type: 'pie',
        radius: ['40%', '70%'],
        center: ['35%', '50%'],
        data: {json.dumps([{"value": v, "name": k} for k, v in signal_freq.items()])},
        label: {{ show: false }},
        emphasis: {{ label: {{ show: true, fontSize: 14, color: '#E8EDF3' }} }}
      }}]
    }});

    // ---- BOROUGH BREAKDOWN ----
    var boroughChart = echarts.init(document.getElementById('boroughChart'));
    var boroughData = {json.dumps(dict(borough_priority))};
    var boroughs = {json.dumps(boroughs)};
    boroughChart.setOption({{
      backgroundColor: '#1A2B45',
      tooltip: {{ trigger: 'axis', axisPointer: {{ type: 'shadow' }}, backgroundColor: '#0F1B2D', borderColor: '#2E4A6E', textStyle: {{ color: '#E8EDF3' }} }},
      legend: {{ data: ['Immediate', 'High', 'Moderate', 'Watchlist'], textStyle: {{ color: '#E8EDF3' }}, top: 5 }},
      xAxis: {{ type: 'category', data: boroughs, axisLabel: {{ color: '#7FB3D3' }}, axisLine: {{ lineStyle: {{ color: '#2E4A6E' }} }} }},
      yAxis: {{ type: 'value', axisLabel: {{ color: '#7FB3D3' }}, splitLine: {{ lineStyle: {{ color: '#2E4A6E' }} }} }},
      series: [
        {{ name: 'Immediate', type: 'bar', stack: 'total', data: boroughs.map(function(b) {{ return (boroughData[b]||{{}})['Immediate']||0; }}), itemStyle: {{ color: '#27AE60' }}, label: {{ show: true, color: '#fff' }} }},
        {{ name: 'High', type: 'bar', stack: 'total', data: boroughs.map(function(b) {{ return (boroughData[b]||{{}})['High']||0; }}), itemStyle: {{ color: '#A9D18E' }}, label: {{ show: true, color: '#000' }} }},
        {{ name: 'Moderate', type: 'bar', stack: 'total', data: boroughs.map(function(b) {{ return (boroughData[b]||{{}})['Moderate']||0; }}), itemStyle: {{ color: '#F1C40F' }}, label: {{ show: true, color: '#000' }} }},
        {{ name: 'Watchlist', type: 'bar', stack: 'total', data: boroughs.map(function(b) {{ return (boroughData[b]||{{}})['Watchlist']||0; }}), itemStyle: {{ color: '#7FB3D3' }}, label: {{ show: true, color: '#000' }} }}
      ]
    }});

    // ---- PRICE vs SCORE ----
    var priceScoreChart = echarts.init(document.getElementById('priceScoreChart'));
    var boroughColors = {{ 'Bronx': '#E74C3C', 'Brooklyn': '#3498DB', 'Manhattan': '#9B59B6', 'Queens': '#1ABC9C' }};
    var psSeries = Object.keys(priceScoreRaw).map(function(b) {{
      return {{
        name: b,
        type: 'scatter',
        data: priceScoreRaw[b],
        symbolSize: 10,
        itemStyle: {{ color: boroughColors[b] || '#7FB3D3', opacity: 0.8 }}
      }};
    }});
    priceScoreChart.setOption({{
      backgroundColor: '#1A2B45',
      tooltip: {{
        trigger: 'item',
        backgroundColor: '#0F1B2D', borderColor: '#2E4A6E', textStyle: {{ color: '#E8EDF3' }},
        formatter: function(p) {{
          return p.seriesName + '<br/>' + p.name + '<br/>Price: $' + (p.value[0]||0).toLocaleString() + '<br/>Score: ' + p.value[1];
        }}
      }},
      legend: {{ data: Object.keys(priceScoreRaw), textStyle: {{ color: '#E8EDF3' }} }},
      xAxis: {{
        type: 'value', name: 'Asking Price ($)',
        nameTextStyle: {{ color: '#7FB3D3' }},
        axisLabel: {{ color: '#7FB3D3', formatter: function(v) {{ return '$' + (v/1000000).toFixed(1) + 'M'; }} }},
        splitLine: {{ lineStyle: {{ color: '#2E4A6E' }} }}
      }},
      yAxis: {{
        type: 'value', name: 'Composite Score',
        nameTextStyle: {{ color: '#7FB3D3' }},
        axisLabel: {{ color: '#7FB3D3' }},
        splitLine: {{ lineStyle: {{ color: '#2E4A6E' }} }}
      }},
      series: psSeries
    }});

    // ---- YEAR BUILT ----
    var yearChart = echarts.init(document.getElementById('yearChart'));
    yearChart.setOption({{
      backgroundColor: '#1A2B45',
      tooltip: {{ trigger: 'axis', backgroundColor: '#0F1B2D', borderColor: '#2E4A6E', textStyle: {{ color: '#E8EDF3' }} }},
      xAxis: {{
        type: 'category',
        data: {json.dumps([str(y) + "s" for y, _ in year_data])},
        axisLabel: {{ color: '#7FB3D3', rotate: 45 }},
        axisLine: {{ lineStyle: {{ color: '#2E4A6E' }} }}
      }},
      yAxis: {{ type: 'value', axisLabel: {{ color: '#7FB3D3' }}, splitLine: {{ lineStyle: {{ color: '#2E4A6E' }} }} }},
      series: [{{
        type: 'bar',
        data: {json.dumps([c for _, c in year_data])},
        itemStyle: {{ color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
          {{ offset: 0, color: '#4FC3F7' }},
          {{ offset: 1, color: '#1F4E79' }}
        ]) }},
        label: {{ show: true, position: 'top', color: '#E8EDF3', fontSize: 10 }}
      }}]
    }});

    // ---- ZONING DISTRIBUTION ----
    var zoneChart = echarts.init(document.getElementById('zoneChart'));
    zoneChart.setOption({{
      backgroundColor: '#1A2B45',
      tooltip: {{ trigger: 'axis', backgroundColor: '#0F1B2D', borderColor: '#2E4A6E', textStyle: {{ color: '#E8EDF3' }} }},
      xAxis: {{
        type: 'category',
        data: {json.dumps([z for z, _ in top_zones])},
        axisLabel: {{ color: '#7FB3D3', rotate: 30 }},
        axisLine: {{ lineStyle: {{ color: '#2E4A6E' }} }}
      }},
      yAxis: {{ type: 'value', axisLabel: {{ color: '#7FB3D3' }}, splitLine: {{ lineStyle: {{ color: '#2E4A6E' }} }} }},
      series: [{{
        type: 'bar',
        data: {json.dumps([c for _, c in top_zones])},
        itemStyle: {{ color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
          {{ offset: 0, color: '#27AE60' }},
          {{ offset: 1, color: '#1A5C38' }}
        ]) }},
        label: {{ show: true, position: 'top', color: '#E8EDF3', fontSize: 10 }}
      }}]
    }});

    // Responsive resize
    window.addEventListener('resize', function() {{
      [geoChart, scoreChart, signalChart, boroughChart, priceScoreChart, yearChart, zoneChart].forEach(function(c) {{ c.resize(); }});
    }});
  </script>
</body>
</html>"""

    output_path = "/tmp/nyc_assemblage/NYC_Assemblage_Dashboard_2026-03-24.html"
    with open(output_path, "w") as f:
        f.write(html)

    print(f"\n✓ Dashboard HTML saved to {output_path}", file=sys.stderr)
    print(json.dumps({"path": output_path}))


if __name__ == "__main__":
    main()
