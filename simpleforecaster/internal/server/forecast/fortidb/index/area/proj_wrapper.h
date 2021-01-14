#include <proj.h>

struct coord {
    double x;
    double y;
};

struct coord convert(PJ *pj, double longitude, double latitude);
