#include "topo.h"
#include "gdal_priv.h"
#include "ogr_spatialref.h"
#include <cmath>
#include <cstdlib>

extern "C"
{
    struct TopoLookup
    {
        GDALDataset *dataset;
        GDALRasterBand *band;
        std::string projection;
        OGRCoordinateTransformation *transform;
        double geotransform[6];
        float noData;
    };

    OGRCoordinateTransformation *topo_transformation(struct TopoLookup *l)
    {
        OGRSpatialReference src;
        if (src.importFromEPSG(4326) != CE_None)
        {
            return nullptr;
        }
        src.SetAxisMappingStrategy(OAMS_TRADITIONAL_GIS_ORDER);

        const char *projection = l->dataset->GetProjectionRef();
        l->projection = projection;
        OGRSpatialReference ref(projection);
        ref.SetAxisMappingStrategy(OAMS_TRADITIONAL_GIS_ORDER);

        return OGRCreateCoordinateTransformation(&src, &ref);
    }

    void topo_init()
    {
        GDALAllRegister();
    }

    struct TopoLookup *topo_open(const char *filename)
    {
        TopoLookup *l = new TopoLookup;

        l->dataset = (GDALDataset *)GDALOpen(filename, GA_ReadOnly);
        if (!l->dataset)
            return nullptr;

        l->band = l->dataset->GetRasterBand(1);
        if (!l->band)
            return nullptr;

        l->noData = l->band->GetNoDataValue();

        l->transform = topo_transformation(l);

        if (l->dataset->GetGeoTransform(l->geotransform) != CE_None)
            return nullptr;

        return l;
    }

    // void topo_close(struct TopoLookup *l)
    // {
    //     OCTDestroyCoordinateTransformation(l->transform);
    //     delete l->dataset;
    //     delete l;
    // }

    const char *topo_projection(struct TopoLookup *l)
    {
        return l->projection.c_str();
    }

    topo_xy topo_transform(struct TopoLookup *l, double latitude, double longitude)
    {
        l->transform->Transform(1, &longitude, &latitude);
        return topo_xy{longitude, latitude};
    }

    int getGridX(struct TopoLookup *l, double x, double y)
    {
        return std::round((x - l->geotransform[0]) / l->geotransform[1]);
    }

    int getGridY(struct TopoLookup *l, double x, double y)
    {
        return std::round((y - l->geotransform[3]) / l->geotransform[5]);
    }

    int topo_contains(struct TopoLookup *l, double x, double y)
    {
        int idxX = getGridX(l, x, y);
        int idxY = getGridY(l, x, y);

        return !(idxX < 0 || idxY < 0 || idxX >= l->band->GetXSize() || idxY >= l->band->GetYSize());
    }

    int topo_lookup(struct TopoLookup *l, float *out, double x, double y)
    {
        int idxX = getGridX(l, x, y);
        int idxY = getGridY(l, x, y);
        if (idxX < 0 || idxY < 0 || idxX >= l->band->GetXSize() || idxY >= l->band->GetYSize())
            return OUT_OF_BOUNDS;

        CPLErr err = l->band->RasterIO(GF_Read, idxX, idxY, 1, 1,
                                       out, 1, 1, GDT_Float32,
                                       0, 0);
        if (err != CE_None)
            return UNABLE_TO_READ_DATA;

        if (*out == l->noData)
            return MISSING_DATA;

        return SUCCESS;
    }
}

const char *topo_mkgrid(struct TopoLookup *l, float *out, int size, double latitude, double longitude)
{
    struct topo_xy xy = topo_transform(l, latitude, longitude);
    int x = getGridX(l, xy.x, xy.y);
    int y = getGridY(l, xy.x, xy.y);

    x -= size / 2;
    y -= size / 2;

    CPLErr err = l->band->RasterIO(GF_Read, x, y, size, size,
                                   out, size, size, GDT_Float32,
                                   0, 0);
    if (err != CE_None)
    {
        return "unable to create image";
    }

    return nullptr;
}
