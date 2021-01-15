#ifndef GEOLOOKUP_H_
#define GEOLOOKUP_H_

#ifdef __cplusplus
extern "C"
{
#endif

struct GridIndex
{
    unsigned Idx;       //< Index into the grid
    unsigned Distance;  //< Result's distance from requested lat/lon, in meters
};

void * MakePointIndex(float * lat, float * lon, unsigned size);
void * EmptyPointIndex();
void Free(void * idx);
void AddCoordinate(void * lookup, float lat, float lon, unsigned value);
struct GridIndex Nearest(void * idx, float latitude, float longitude);

#ifdef __cplusplus
}
#endif

#endif