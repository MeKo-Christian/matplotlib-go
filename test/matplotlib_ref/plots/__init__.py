from __future__ import annotations

import importlib

PLOT_NAMES = ['basic_line', 'basic_line_labels', 'joins_caps', 'dashes', 'line2d_markers', 'scatter_basic', 'scatter_marker_types', 'scatter_advanced', 'bar_basic_frame', 'bar_basic_ticks', 'bar_basic_tick_labels', 'bar_basic_title', 'bar_basic', 'bar_horizontal', 'bar_grouped', 'fill_basic', 'fill_between', 'fill_stacked', 'errorbar_basic', 'multi_series_basic', 'multi_series_color_cycle', 'hist_basic', 'hist_density', 'hist_strategies', 'boxplot_basic', 'text_labels_strict', 'title_strict', 'mathtext_basic', 'mathtext_fractions', 'mathtext_integrals', 'mathtext_matrices', 'mathtext_inline_labels', 'image_heatmap', 'imshow_clipped', 'imshow_transformed', 'image_alpha', 'matshow_basic', 'spy_marker', 'spy_image', 'named_colors', 'colormap_diverging', 'colormap_qualitative', 'colormap_cyclic', 'axes_top_right_inverted', 'axes_control_surface', 'axes_option_breadth', 'transform_coordinates', 'gridspec_composition', 'figure_labels_composition', 'colorbar_composition', 'annotation_composition', 'patch_showcase', 'mesh_contour_tri', 'plot_variants', 'spectrum_variants', 'stat_variants', 'specialty_depth', 'stem_plot', 'specialty_artists', 'units_overview', 'units_dates', 'units_categories', 'units_custom_converter', 'vector_fields', 'polar_axes', 'geo_mollweide_axes', 'geo_aitoff_axes', 'geo_hammer_axes', 'geo_lambert_axes', 'radar_basic', 'skewt_basic', 'mplot3d_basic', 'mplot3d_terrain', 'mplot3d_plot3d', 'mplot3d_scatter3d', 'mplot3d_surface3d', 'mplot3d_wire3d', 'mplot3d_trisurf3d', 'mplot3d_bar3d', 'mplot3d_voxels', 'mplot3d_quiver3d', 'mplot3d_errorbar3d', 'mplot3d_stem3d', 'mplot3d_fill_between3d', 'unstructured_showcase', 'arrays_showcase', 'axisartist_showcase', 'axes_grid1_showcase', 'pcolor_flat', 'pcolormesh_nearest', 'pcolormesh_gouraud', 'pcolormesh_masked', 'hist2d_weighted_density', 'asinh_norm_image', 'boundarynorm_pcolormesh', 'collection_mutable_scalarmap', 'colorbar_boundary_values', 'colorbar_horizontal_ticks', 'lognorm_imshow', 'twoslope_norm_image', 'colorbar_extensions', 'large_scatter', 'mixed_collection', 'quad_mesh', 'gouraud_triangles', 'clip_path_batch']

def load_plot(name: str):
    if name not in PLOT_NAMES:
        raise KeyError(name)
    module = importlib.import_module(f"{__name__}.{name}")
    return module.PLOT

def all_plots():
    return [load_plot(name) for name in PLOT_NAMES]
