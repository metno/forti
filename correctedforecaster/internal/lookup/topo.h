#ifdef __cplusplus
extern "C"
{
#endif

    struct TopoLookup;

    void topo_init();

    struct TopoLookup *topo_open(const char *filename);

    // void topo_close(struct TopoLookup *l);

    const char *topo_projection(struct TopoLookup *l);

    struct topo_xy
    {
        double x;
        double y;
    };

    struct topo_xy topo_transform(struct TopoLookup *l, double latitude, double longitude);

    int topo_contains(struct TopoLookup *l, double x, double y);

    int topo_lookup(struct TopoLookup *l, float *out, double x, double y);

    const char *topo_mkgrid(struct TopoLookup *l, float *out, int size, double latitude, double longitude);

    const int SUCCESS = 0;
    const int OUT_OF_BOUNDS = 1;
    const int MISSING_DATA = 2;
    const int UNABLE_TO_READ_DATA = 3;

#ifdef __cplusplus
}
#endif
