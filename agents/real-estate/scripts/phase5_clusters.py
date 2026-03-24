#!/usr/bin/env python3
"""
Phase 5: Cluster Detection
Identify blocks where 2+ qualifying R7+ properties are present
"""

import json
import sys
from collections import defaultdict


def main():
    with open("/tmp/nyc_assemblage/final_properties.json") as f:
        properties = json.load(f)

    # Group by (borough, block)
    block_groups = defaultdict(list)

    for prop in properties:
        borough = prop.get("_borough", "")
        block = prop.get("_block", "")
        if borough and block:
            block_groups[(borough, block)].append(prop)

    # Find clusters (2+ properties on same block)
    clusters = []
    for (borough, block), props in block_groups.items():
        if len(props) >= 2:
            # Calculate cluster metrics
            total_lot_area = 0
            total_ask = 0
            addresses = []
            scores = []
            zpids = []
            zones = set()

            for p in props:
                try:
                    lot_area = int(float(p.get("_lot_area", 0) or 0))
                    total_lot_area += lot_area
                except (ValueError, TypeError):
                    pass

                price = p.get("price", 0) or 0
                total_ask += price

                addresses.append(p.get("address", ""))
                scores.append(p.get("_score", 0))
                zpids.append(p.get("zpid", ""))
                zones.add(p.get("_zoning", ""))

            cluster = {
                "borough": borough,
                "block": block,
                "property_count": len(props),
                "addresses": addresses,
                "zpids": zpids,
                "total_lot_area_sf": total_lot_area if total_lot_area > 0 else "Unable to verify",
                "combined_asking_price": total_ask if total_ask > 0 else "Unable to verify",
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
    cluster_zpids = set()
    for cluster in clusters:
        for zpid in cluster["zpids"]:
            cluster_zpids.add(zpid)

    for prop in properties:
        prop["_in_cluster"] = prop.get("zpid", "") in cluster_zpids
        if prop["_in_cluster"]:
            # Find cluster block
            for cluster in clusters:
                if prop.get("zpid", "") in cluster["zpids"]:
                    prop["_cluster_size"] = cluster["property_count"]
                    prop["_cluster_combined_ask"] = cluster["combined_asking_price"]
                    prop["_cluster_total_lot"] = cluster["total_lot_area_sf"]
                    break

            # Add score bonus for adjacent lot also for sale
            if not prop.get("_cluster_bonus_applied"):
                prop["_score"] = prop.get("_score", 0) + 4
                prop["_score_reasons"] = prop.get("_score_reasons", []) + ["Adjacent lot also for sale (+4)"]
                prop["_cluster_bonus_applied"] = True

                # Recompute priority
                score = prop["_score"]
                if score >= 20:
                    prop["_priority"] = "Immediate"
                elif score >= 15:
                    prop["_priority"] = "High"
                elif score >= 10:
                    prop["_priority"] = "Moderate"
                else:
                    prop["_priority"] = "Watchlist"

    # Re-sort by score
    properties.sort(key=lambda p: p.get("_score", 0), reverse=True)

    print(f"=== Cluster Detection Results ===", file=sys.stderr)
    print(f"  Total clusters (2+ properties same block): {len(clusters)}", file=sys.stderr)
    print(f"  Properties in clusters: {len(cluster_zpids)}", file=sys.stderr)
    print(f"\n  Top 5 clusters:", file=sys.stderr)
    for c in clusters[:5]:
        print(f"    {c['borough']} Block {c['block']}: {c['property_count']} properties, "
              f"combined ask ${c['combined_asking_price']:,.0f}" if isinstance(c['combined_asking_price'], (int, float)) else
              f"    {c['borough']} Block {c['block']}: {c['property_count']} properties",
              file=sys.stderr)
        for addr in c["addresses"]:
            print(f"      - {addr}", file=sys.stderr)

    # Save
    with open("/tmp/nyc_assemblage/final_properties.json", "w") as f:
        json.dump(properties, f, indent=2)

    with open("/tmp/nyc_assemblage/clusters.json", "w") as f:
        # Save clusters without the full property objects to keep file manageable
        slim_clusters = []
        for c in clusters:
            slim = {k: v for k, v in c.items() if k != "properties"}
            slim_clusters.append(slim)
        json.dump(slim_clusters, f, indent=2)

    print(f"\n✓ Saved {len(clusters)} clusters and updated property scores", file=sys.stderr)
    print(json.dumps({"clusters": len(clusters), "clustered_properties": len(cluster_zpids)}))


if __name__ == "__main__":
    main()
