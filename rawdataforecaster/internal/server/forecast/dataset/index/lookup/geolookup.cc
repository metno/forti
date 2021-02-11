#include "geolookup.h"
#include <limits>
#include <map>
#include <s2/s2point_index.h>
#include <s2/s2earth.h>
#include <s2/s2latlng.h>
#include <s2/s2point.h>
#include <s2/s2closest_point_query.h>

extern "C"
{
void * MakePointIndex(float * lat, float * lon, unsigned size)
{
    S2PointIndex<unsigned> * ret = new S2PointIndex<unsigned>;
    for (unsigned i = 0; i < size; ++i)
    {
		S2Point point(S2LatLng::FromDegrees(lat[i], lon[i]));
		ret->Add(point, i);
    }
    return (void*) ret;
}

void * EmptyPointIndex()
{
	return (void*) new S2PointIndex<unsigned>;
}

void Free(void * idx)
{
    delete (S2PointIndex<unsigned> *) idx;
}

void AddCoordinate(void * lookup, float lat, float lon, unsigned value)
{
	S2PointIndex<unsigned>* idx = (S2PointIndex<unsigned>*) lookup;
	S2Point point(S2LatLng::FromDegrees(lat, lon));
	idx->Add(point, value);

}

GridIndex Nearest(void * lookup, float latitude, float longitude)
{
    S2PointIndex<unsigned>* idx = (S2PointIndex<unsigned>*) lookup;

	S2Point point(S2LatLng::FromDegrees(latitude, longitude));
	S2ClosestPointQuery<unsigned> closestPointQuery(idx);
	S2ClosestPointQuery<unsigned>::PointTarget target(point);

	S2ClosestPointQuery<unsigned>::Result result = closestPointQuery.FindClosestPoint(&target);
	if (result.is_empty())
        return GridIndex{0, std::numeric_limits<unsigned>::max()};

	unsigned distance = S2Earth::RadiansToMeters(result.distance().radians());
	S2LatLng point = S2LatLng(result.point());

	return GridIndex{result.data(), distance, point};

}
}
