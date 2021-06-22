#include "proj_wrapper.h"


struct coord convert(PJ *pj, double longitude, double latitude) {
    // since PJ_COORD is a union, it is hard to implement this clearly in pure go.
    PJ_COORD coord = proj_coord(proj_torad(longitude), proj_torad(latitude), 0, 0);

    coord = proj_trans(pj, PJ_FWD, coord);

    if (proj_angular_output(pj, PJ_FWD)) {
        coord.xy.x = proj_todeg(coord.xy.x);
        coord.xy.y = proj_todeg(coord.xy.y);
    }

    struct coord c;
    c.x = coord.xy.x;
    c.y = coord.xy.y;
    return c;
}
