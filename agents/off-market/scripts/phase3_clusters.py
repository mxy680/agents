#!/usr/bin/env python3
"""
Phase 3: Cluster Detection (Off-Market R8+ Properties)
Identify blocks where 2+ qualifying R8+ properties are present.

Input:  /tmp/off_market_scan/scored_properties.json
Output: /tmp/off_market_scan/final_properties.json
        /tmp/off_market_scan/clusters.json
"""

import json
import sys
from collections import defaultdict


def main():
    with open("/tmp/off_market_scan/scored_properties.json") as f:
        properties = json.load(f)

    # Group by (borough, block)
    block_groups: dict = defaultdict(list)

    for prop in properties:
        borough = prop.get("_borough", "")
        block = prop.get("_block", "")
        if borough and block:
            block_groups[(borough, block)].append(prop)

    # Find clusters (2+ properties on same block)
    clusters = []
    for (borough, block), props in block_groups.items():
        if len(props) >= 2:
            total_lot_area = 0
            addresses = []
            scores = []
            bbls = []
            zones: set = set()

            for p in props:
                try:
                    lot_area = int(float(str(p.get("_lot_area", 0) or 0).replace(",", "").replace("$", "")))
                    total_lot_area += lot_area
                except (ValueError, TypeError):
                    pass

                addresses.append(p.get("address", ""))
                scores.append(p.get("_score", 0))
                bbls.append(p.get("_bbl", ""))
                zones.add(p.get("_zoning", ""))

            cluster = {
                "borough": borough,
                "block": block,
                "property_count": len(props),
                "addresses": addresses,
                "bbls": bbls,
                "total_lot_area_sf": total_lot_area if total_lot_area > 0 else "Unable to verify",
                "zones": list(zones),
                "max_score": max(scores),
                "avg_score": round(sum(scores) / len(scores), 1),
                "combined_score": sum(scores),
                "properties": props
            }
            clusters.append(cluster)

    # Sort clusters by combined score
    clusters.sort(key=lambda c: c["combined_score"], reverse=True)

    # Tag properties that are in clusters
    cluster_bbls: set = set()
    for cluster in clusters:
        for bbl in cluster["bbls"]:
            cluster_bbls.add(bbl)

    for prop in properties:
        prop["_in_cluster"] = prop.get("_bbl", "") in cluster_bbls
        if prop["_in_cluster"]:
            for cluster in clusters:
                if prop.get("_bbl", "") in cluster["bbls"]:
                    prop["_cluster_size"] = cluster["property_count"]
                    prop["_cluster_total_lot"] = cluster["total_lot_area_sf"]
                    break

            # Add score bonus for adjacent off-market lot on same block
            if not prop.get("_cluster_bonus_applied"):
                prop["_score"] = prop.get("_score", 0) + 4
                prop["_score_reasons"] = prop.get("_score_reasons", []) + ["Adjacent off-market lot on same block (+4)"]
                prop["_cluster_bonus_applied"] = True

                # Recompute priority
                score = prop["_score"]
                if score >= 15:
                    prop["_priority"] = "Immediate"
                elif score >= 10:
                    prop["_priority"] = "High"
                elif score >= 6:
                    prop["_priority"] = "Moderate"
                else:
                    prop["_priority"] = "Watchlist"

    # Re-sort by score
    properties.sort(key=lambda p: p.get("_score", 0), reverse=True)

    print(f"=== Cluster Detection Results ===", file=sys.stderr)
    print(f"  Total clusters (2+ properties same block): {len(clusters)}", file=sys.stderr)
    print(f"  Properties in clusters: {len(cluster_bbls)}", file=sys.stderr)
    print(f"\n  Top 5 clusters:", file=sys.stderr)
    for c in clusters[:5]:
        print(f"    {c['borough']} Block {c['block']}: {c['property_count']} properties",
              file=sys.stderr)
        for addr in c["addresses"]:
            print(f"      - {addr}", file=sys.stderr)

    # Save
    with open("/tmp/off_market_scan/final_properties.json", "w") as f:
        json.dump(properties, f, indent=2)

    with open("/tmp/off_market_scan/clusters.json", "w") as f:
        slim_clusters = []
        for c in clusters:
            slim = {k: v for k, v in c.items() if k != "properties"}
            slim_clusters.append(slim)
        json.dump(slim_clusters, f, indent=2)

    print(f"\nSaved {len(clusters)} clusters and updated property scores", file=sys.stderr)
    print(json.dumps({"clusters": len(clusters), "clustered_properties": len(cluster_bbls)}))


if __name__ == "__main__":
    main()
