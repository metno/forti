#include "proj_wrapper.h"


static double deg2rad(double deg)
{
    return deg / (180.0f / 3.14159265359f);
}


struct coord convert(PJ *pj, double longitude, double latitude) {
    // since PJ_COORD is a union, it is hard to implement this clearly in pure go.
    PJ_COORD coord = proj_coord(deg2rad(longitude), deg2rad(latitude), 0, 0);

    coord = proj_trans(pj, PJ_FWD, coord);

    struct coord c;
    c.x = coord.xy.x;
    c.y = coord.xy.y;
    return c;
}
