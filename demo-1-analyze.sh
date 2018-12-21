#!/bin/bash

video_file_path=$(realpath $1)
nb_frames=$(ffprobe -v error -select_streams v:0 -show_entries stream=nb_frames -of default=nokey=1:noprint_wrappers=1 $1)

## Create input.csv
echo 'org_video,label,start_frm,video_id' > ../my_list.csv
for i in $(seq 1 10);
do
  echo ${video_file_path},0,$(($nb_frames / 10 * ($i - 1))),${i} >> ../my_list.csv
done

## Create DB
rm -rf $HOME/my_lmdb_data
LD_LIBRARY_PATH=$HOME/local/lib PYTHONPATH=$(pwd)/lib python data/create_video_db.py \
--list_file=$HOME/my_list.csv \
--output_file=$HOME/my_lmdb_data \
--use_list=1 --use_video_id=1 --use_start_frame=1

## Extract features
# Use 'nvidia-smi' to find vacant GPUs
LD_LIBRARY_PATH=$HOME/local/lib PYTHONPATH=$HOME/pytorch/build:$(pwd)/lib python tools/extract_features.py \
--test_data=$HOME/my_lmdb_data \
--model_name=r2plus1d --model_depth=34 --clip_length_rgb=32 \
--gpus=0,1,2,3,4 \
--batch_size=2 \
--load_model_path=$HOME/r2.5d_d34_l32.pkl \
--output_path=$HOME/my_features.pkl \
--features=softmax,final_avg,video_id \
--sanity_check=0 --get_video_id=1 --use_local_file=1 --num_labels=400 --num_iterations=1
