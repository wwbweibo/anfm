import logging
import torch
import cn_clip.clip as clip
from cn_clip.clip import load_from_name
from lavis.models import load_model_and_preprocess
from models.image import ImageUnderstandingResult, ImageLabel
from PIL import Image
import importlib.util as importutil
import numpy as np

class ImageUnderstanding:
    def __init__(self):
        # check if torch_directml is available
        if importutil.find_spec("torch_directml") is None:
            self.device = torch.device("cuda" if torch.cuda.is_available() else "cpu")
        else:
            import torch_directml
            self.device = torch_directml.device()
        self.clip_model = None
        self.clip_preprocess = None
        self.text_labels = None
        self.text_feature = None
        self.caption_model = None
        self.caption_vis_processors = None
        self.__init_clip_model__()
        self.__init_caption_model__()

    def __init_clip_model__(self):
        self.clip_model, self.clip_preprocess = load_from_name("ViT-B-16", device=self.device, download_root="./")
        text_labels = []
        with open("label_cn.txt") as f:
            text_labels = f.readlines()
        # remove whitespace characters like `\n` at the end of each line
        text_labels = [x.strip() for x in text_labels]
        self.text_labels = text_labels
        labels = clip.tokenize(text_labels).to(self.device)
        with torch.no_grad():
            labels = clip.tokenize(text_labels).to(self.device)
            text_feature = self.clip_model.encode_text(labels)
            text_feature /= text_feature.norm(dim=-1, keepdim=True)
            self.text_feature = text_feature

    def __init_caption_model__(self):
        self.caption_model, self.caption_vis_processors, _  = load_model_and_preprocess('blip_caption', model_type='base_coco', is_eval=True,  device=self.device)

    def label_image(self, path: str) -> list[ImageLabel]:
        if self.clip_model is None:
            self.__init_clip_model__()
        pil_image = Image.open(path)
        # 检查图像通道，如果为4通道，去掉alpha通道
        if pil_image.mode == "RGBA":
            pil_image = pil_image.convert("RGB")
        image = self.clip_preprocess(pil_image).unsqueeze(0).to(self.device)
        with torch.no_grad():
            image_features = self.clip_model.encode_image(image)
            image_features /= image_features.norm(dim=-1, keepdim=True)
        similarity = (100.0 * image_features @ self.text_feature.T).softmax(dim=-1)
        # find top 5 labels with highest similarity
        values, indices = similarity[0].topk(5)
        # print("Label and confidence:")
        predictions = []
        for value, index in zip(values, indices):
            predictions.append(ImageLabel(self.text_labels[index], value.item()))
        return predictions

    def caption_image(self, path: str) -> str:
        if self.caption_model is None:
            self.__init_caption_model__()
        pil_image = Image.open(path)
        # 检查图像通道，如果为4通道，去掉alpha通道
        if pil_image.mode == "RGBA":
            pil_image = pil_image.convert("RGB")
        return self.caption_model.generate({"image": self.caption_vis_processors['eval'](pil_image).unsqueeze(0).to(self.device)})[0]

    def understand(self, path: str) -> ImageUnderstandingResult:
        labels = self.label_image(path)
        logging.info("Image Labels: %s", labels)
        caption = self.caption_image(path)
        logging.info("Image Caption: %s", caption)
        return ImageUnderstandingResult(labels, caption)
